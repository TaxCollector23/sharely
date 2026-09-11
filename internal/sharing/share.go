// Package sharing owns the Share domain object and its lifecycle: creation,
// expiration, activity tracking, and the in-memory registry that the server
// and CLI both operate on.
package sharing

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type Type string

const (
	TypeFile      Type = "file"
	TypeDirectory Type = "directory"
	TypeWebsite   Type = "website"
	TypeProxy     Type = "proxy"
)

type Status string

const (
	StatusRunning Status = "running"
	StatusExpired Status = "expired"
	StatusStopped Status = "stopped"
)

// Duration is a named, human-facing expiration choice.
type Duration string

const (
	Duration15Min   Duration = "15m"
	Duration1Hour   Duration = "1h"
	Duration4Hour   Duration = "4h"
	DurationForever Duration = "forever"
)

func (d Duration) TTL() (time.Duration, bool) {
	switch d {
	case Duration15Min:
		return 15 * time.Minute, true
	case Duration1Hour:
		return time.Hour, true
	case Duration4Hour:
		return 4 * time.Hour, true
	case DurationForever:
		return 0, false
	default:
		return time.Hour, true
	}
}

func (d Duration) Label() string {
	switch d {
	case Duration15Min:
		return "15 minutes"
	case Duration1Hour:
		return "1 hour"
	case Duration4Hour:
		return "4 hours"
	case DurationForever:
		return "Until I stop it"
	default:
		return string(d)
	}
}

// visitor tracks a distinct device so the dashboard can show "N devices
// connected" without keeping invasive logs.
type visitor struct {
	key      string
	lastSeen time.Time
}

type Share struct {
	mu sync.RWMutex

	ID         string
	TargetArg  string // the raw argument the user passed, for "share again"
	RootDir    string // absolute, symlink-resolved directory actually served
	EntryPoint string // relative path to the site's entry file, if any
	Type       Type

	ProxyUpstream string // e.g. http://127.0.0.1:5173, for TypeProxy

	CreatedAt time.Time
	ExpiresAt *time.Time // nil == until stopped
	Duration  Duration

	PasswordHash string
	HasPassword  bool

	status Status

	visitors    map[string]*visitor
	lastAccess  *time.Time
	recentLogs  []AccessLog
	maxLogLines int
}

type AccessLog struct {
	Time   time.Time `json:"time"`
	Method string    `json:"method"`
	Path   string    `json:"path"`
	Status int       `json:"status"`
}

func newID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *Share) Status() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.effectiveStatusLocked()
}

func (s *Share) effectiveStatusLocked() Status {
	if s.status == StatusStopped {
		return StatusStopped
	}
	if s.ExpiresAt != nil && time.Now().After(*s.ExpiresAt) {
		return StatusExpired
	}
	return StatusRunning
}

func (s *Share) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = StatusStopped
}

// SetExpiration changes the share's expiration without altering its ID,
// URL, or any other identity — exactly the "Change time" UX requirement.
func (s *Share) SetExpiration(d Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Duration = d
	if ttl, hasTTL := d.TTL(); hasTTL {
		exp := time.Now().Add(ttl)
		s.ExpiresAt = &exp
	} else {
		s.ExpiresAt = nil
	}
}

func (s *Share) RemainingLabel() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	switch s.effectiveStatusLocked() {
	case StatusStopped:
		return "Sharing stopped"
	case StatusExpired:
		return "This share has ended"
	}
	if s.ExpiresAt == nil {
		return "Available until stopped"
	}
	remaining := time.Until(*s.ExpiresAt)
	if remaining < 0 {
		return "This share has ended"
	}
	return "Available for " + humanDuration(remaining)
}

func humanDuration(d time.Duration) string {
	if d < time.Minute {
		return "less than a minute"
	}
	mins := int(d.Minutes())
	if mins < 60 {
		if mins == 1 {
			return "1 minute"
		}
		return itoa(mins) + " minutes"
	}
	hours := mins / 60
	rem := mins % 60
	label := itoa(hours) + " hour"
	if hours != 1 {
		label += "s"
	}
	if rem > 0 {
		label += " " + itoa(rem) + "m"
	}
	return label
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// Touch records activity from a distinct visitor (keyed by a hash of IP +
// user agent, never stored raw) and appends a lightweight access log line.
func (s *Share) Touch(visitorKey, method, path string, status int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	s.lastAccess = &now
	if s.visitors == nil {
		s.visitors = make(map[string]*visitor)
	}
	s.visitors[visitorKey] = &visitor{key: visitorKey, lastSeen: now}
	// Prune visitors idle for more than 5 minutes so the count reflects who
	// is actually still around, not everyone who ever connected.
	for k, v := range s.visitors {
		if now.Sub(v.lastSeen) > 5*time.Minute {
			delete(s.visitors, k)
		}
	}

	if s.maxLogLines == 0 {
		s.maxLogLines = 200
	}
	s.recentLogs = append(s.recentLogs, AccessLog{Time: now, Method: method, Path: path, Status: status})
	if len(s.recentLogs) > s.maxLogLines {
		s.recentLogs = s.recentLogs[len(s.recentLogs)-s.maxLogLines:]
	}
}

func (s *Share) DeviceCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.visitors)
}

func (s *Share) LastAccess() *time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastAccess
}

func (s *Share) Logs() []AccessLog {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]AccessLog, len(s.recentLogs))
	copy(out, s.recentLogs)
	return out
}
