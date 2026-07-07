package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"probe-shield/internal/config"
)

const SessionCookieName = "probe_shield_session"

type Service struct {
	login      string
	password   string
	sessionTTL time.Duration
	limiter    *RateLimiter

	mu       sync.Mutex
	sessions map[string]time.Time
}

func NewService(cfg config.Config) *Service {
	return &Service{
		login:      cfg.Auth.Login,
		password:   cfg.Auth.Password,
		sessionTTL: cfg.Auth.SessionTTL,
		limiter:    NewRateLimiter(cfg.RateLimit.Window, cfg.RateLimit.MaxFailedAttempts),
		sessions:   make(map[string]time.Time),
	}
}

func (s *Service) HasCredentials() bool {
	return s.login != "" && s.password != ""
}

func (s *Service) IsBlocked(r *http.Request) (bool, time.Duration) {
	return s.limiter.IsBlocked(clientKey(r))
}

func (s *Service) TryLogin(r *http.Request, login, password string) (ok bool, blocked bool, retryAfter time.Duration) {
	key := clientKey(r)
	if isBlocked, retry := s.limiter.IsBlocked(key); isBlocked {
		return false, true, retry
	}

	if s.HasCredentials() && constantTimeEqual(login, s.login) && constantTimeEqual(password, s.password) {
		s.limiter.Reset(key)
		return true, false, 0
	}

	isBlocked, retry := s.limiter.RecordFailure(key)
	return false, isBlocked, retry
}

func (s *Service) CreateSession(w http.ResponseWriter, secure bool) (string, error) {
	token, err := newSessionToken()
	if err != nil {
		return "", err
	}
	expiresAt := time.Now().Add(s.sessionTTL)

	s.mu.Lock()
	s.sessions[token] = expiresAt
	s.mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   int(s.sessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	return token, nil
}

func (s *Service) Authenticated(r *http.Request) bool {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil || cookie.Value == "" {
		return false
	}

	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	expiresAt, ok := s.sessions[cookie.Value]
	if !ok {
		return false
	}
	if now.After(expiresAt) {
		delete(s.sessions, cookie.Value)
		return false
	}
	return true
}

func (s *Service) DestroySession(w http.ResponseWriter, r *http.Request, secure bool) {
	if cookie, err := r.Cookie(SessionCookieName); err == nil {
		s.mu.Lock()
		delete(s.sessions, cookie.Value)
		s.mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func constantTimeEqual(a, b string) bool {
	ah := sha256.Sum256([]byte(a))
	bh := sha256.Sum256([]byte(b))
	return subtle.ConstantTimeCompare(ah[:], bh[:]) == 1
}

func newSessionToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func clientKey(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}
