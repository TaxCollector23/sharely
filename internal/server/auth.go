package server

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/TaxCollector23/sharely/internal/config"
	"github.com/TaxCollector23/sharely/internal/sharing"

	"golang.org/x/crypto/bcrypt"
)

// authSecret is generated fresh per daemon process; session cookies never
// need to survive a restart since the shares themselves don't either.
var authSecret = randomSecret()

func randomSecret() []byte {
	b := make([]byte, 32)
	rand.Read(b)
	return b
}

func signToken(shareID string, expiry time.Time) string {
	payload := shareID + "." + strconv.FormatInt(expiry.Unix(), 10)
	mac := hmac.New(sha256.New, authSecret)
	mac.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return payload + "." + sig
}

func verifyToken(shareID, token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}
	payload := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, authSecret)
	mac.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(parts[2])) {
		return false
	}
	if parts[0] != shareID {
		return false
	}
	expUnix, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return false
	}
	return time.Now().Before(time.Unix(expUnix, 0))
}

func cookieName(shareID string) string {
	return config.SessionCookiePrefix + shareID
}

func isAuthenticated(r *http.Request, s *sharing.Share) bool {
	if !s.HasPassword {
		return true
	}
	c, err := r.Cookie(cookieName(s.ID))
	if err != nil {
		return false
	}
	return verifyToken(s.ID, c.Value)
}

func setAuthCookie(w http.ResponseWriter, s *sharing.Share) {
	expiry := time.Now().Add(24 * time.Hour)
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName(s.ID),
		Value:    signToken(s.ID, expiry),
		Path:     "/" + s.ID,
		Expires:  expiry,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func checkPassword(s *sharing.Share, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(s.PasswordHash), []byte(password)) == nil
}

func renderPasswordPage(w http.ResponseWriter, shareID string, wrong bool) {
	writeHTML(w, http.StatusUnauthorized, passwordPage(shareID, wrong))
}
