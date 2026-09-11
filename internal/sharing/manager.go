package sharing

import (
	"fmt"
	"sync"
	"time"

	"sharely/internal/names"

	"golang.org/x/crypto/bcrypt"
)

// Manager is the single in-memory registry of every share created by this
// daemon process. It is safe for concurrent use by the content server, the
// control API, and the expiration sweeper.
type Manager struct {
	mu     sync.RWMutex
	shares map[string]*Share
}

func NewManager() *Manager {
	return &Manager{shares: make(map[string]*Share)}
}

type CreateOptions struct {
	TargetArg     string
	RootDir       string
	EntryPoint    string
	Type          Type
	ProxyUpstream string
	Duration      Duration
	Password      string
}

func (m *Manager) Create(opts CreateOptions) (*Share, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	taken := make(map[string]bool, len(m.shares))
	for id := range m.shares {
		taken[id] = true
	}
	id, err := names.Generate(taken)
	if err != nil {
		return nil, err
	}

	s := &Share{
		ID:            id,
		TargetArg:     opts.TargetArg,
		RootDir:       opts.RootDir,
		EntryPoint:    opts.EntryPoint,
		Type:          opts.Type,
		ProxyUpstream: opts.ProxyUpstream,
		CreatedAt:     time.Now(),
		status:        StatusRunning,
	}

	dur := opts.Duration
	if dur == "" {
		dur = Duration1Hour
	}
	s.SetExpiration(dur)

	if opts.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(opts.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("hashing password: %w", err)
		}
		s.PasswordHash = string(hash)
		s.HasPassword = true
	}

	m.shares[id] = s
	return s, nil
}

func (m *Manager) Get(id string) (*Share, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.shares[id]
	return s, ok
}

// List returns every share, most recently created first.
func (m *Manager) List() []*Share {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Share, 0, len(m.shares))
	for _, s := range m.shares {
		out = append(out, s)
	}
	// newest first
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].CreatedAt.After(out[i].CreatedAt) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

// Active returns only shares that are currently running (not expired or stopped).
func (m *Manager) Active() []*Share {
	var out []*Share
	for _, s := range m.List() {
		if s.Status() == StatusRunning {
			out = append(out, s)
		}
	}
	return out
}

func (m *Manager) Stop(id string) bool {
	m.mu.RLock()
	s, ok := m.shares[id]
	m.mu.RUnlock()
	if !ok {
		return false
	}
	s.Stop()
	return true
}

func (m *Manager) StopAll() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n := 0
	for _, s := range m.shares {
		if s.Status() == StatusRunning {
			s.Stop()
			n++
		}
	}
	return n
}

// Count reports how many shares are registered (any status) — used to
// decide whether this process should keep running.
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.shares)
}

// Sweep clears out shares that have been stopped/expired long enough that
// they no longer need to occupy the registry (housekeeping only; the ID
// stays reserved for the process lifetime otherwise memory is trivial).
func (m *Manager) Sweep(olderThan time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, s := range m.shares {
		if s.Status() == StatusRunning {
			continue
		}
		if time.Since(s.CreatedAt) > olderThan {
			delete(m.shares, id)
		}
	}
}
