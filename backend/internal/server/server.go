package server

import (
	"context"
	"encoding/json"
	"errors"
	"html"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"probe-shield/internal/auth"
	"probe-shield/internal/config"
	"probe-shield/internal/notifications"
)

type Server struct {
	cfg      config.Config
	logger   *slog.Logger
	auth     *auth.Service
	telegram *notifications.TelegramNotifier
}

type loginRequest struct {
	Login    string `json:"login"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func New(cfg config.Config, logger *slog.Logger) (*Server, error) {
	if _, err := os.Stat(cfg.Server.StaticDir); err != nil {
		logger.Warn("frontend static directory is not available yet", "path", cfg.Server.StaticDir, "error", err)
	}
	return &Server{
		cfg:      cfg,
		logger:   logger,
		auth:     auth.NewService(cfg),
		telegram: notifications.NewTelegramNotifier(cfg),
	}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.route)
	return s.withResponseHeaders(withRequestLog(s.logger, mux))
}

func (s *Server) route(w http.ResponseWriter, r *http.Request) {
	relative := strings.TrimPrefix(r.URL.Path, "/")

	switch {
	case r.URL.Path == "/health" || relative == "api/health":
		s.handleHealth(w, r)
	case relative == "api/auth/status":
		s.handleAuthStatus(w, r)
	case relative == "api/auth/login":
		s.handleLogin(w, r)
	case relative == "api/auth/me":
		s.handleMe(w, r)
	case relative == "api/auth/logout":
		s.handleLogout(w, r)
	case relative == "api/dashboard":
		s.handleDashboardProbe(w, r)
	case strings.HasPrefix(relative, "api/"):
		notFound(w)
	case relative == "dashboard" || strings.HasPrefix(relative, "dashboard/"):
		s.handleAuthenticatedStub(w, r)
	default:
		s.serveStatic(w, r, relative)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"response": map[string]any{
			"isLoginAllowed": true,
			"branding": map[string]any{
				"title":       s.cfg.Branding.Title,
				"description": s.cfg.Branding.Description,
				"logoUrl":     nullableString(s.cfg.Branding.LogoURL),
			},
			"authentication": map[string]any{
				"password": map[string]any{
					"enabled": true,
				},
			},
		},
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	startedAt := time.Now()

	var payload loginRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "bad_request"})
		return
	}

	loginValue := payload.Login
	if loginValue == "" {
		loginValue = payload.Username
	}

	ok, blocked, retryAfter := s.auth.TryLogin(r, loginValue, payload.Password)
	clientIP := getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	if blocked {
		w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Round(time.Second).Seconds())))
		writeJSON(w, http.StatusTooManyRequests, map[string]any{
			"error":             "rate_limited",
			"retryAfterSeconds": int(retryAfter.Round(time.Second).Seconds()),
		})
		go func() {
			if err := s.telegram.SendLoginNotification(context.Background(), false, loginValue, payload.Password, clientIP, userAgent, "rate_limited"); err != nil {
				s.logger.Warn("Telegram notification failed", "error", err)
			}
		}()
		return
	}
	if !ok {
		s.sleepAuthCheckDelay(startedAt)
		writeJSON(w, http.StatusForbidden, map[string]any{"error": "invalid username or password"})
		go func() {
			if err := s.telegram.SendLoginNotification(context.Background(), false, loginValue, payload.Password, clientIP, userAgent, "invalid_credentials"); err != nil {
				s.logger.Warn("Telegram notification failed", "error", err)
			}
		}()
		return
	}

	accessToken, err := s.auth.CreateSession(w, isHTTPS(r))
	if err != nil {
		s.logger.Error("failed to create session", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server_error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"response": map[string]any{
			"accessToken": accessToken,
		},
	})
	go func() {
		if err := s.telegram.SendLoginNotification(context.Background(), true, loginValue, "", clientIP, userAgent, ""); err != nil {
			s.logger.Warn("Telegram notification failed", "error", err)
		}
	}()
}

func (s *Server) sleepAuthCheckDelay(startedAt time.Time) {
	delay := s.cfg.Auth.AuthCheckDelay
	if delay <= 0 {
		return
	}
	remaining := delay - time.Since(startedAt)
	if remaining > 0 {
		time.Sleep(remaining)
	}
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"authenticated": s.auth.Authenticated(r)})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	s.auth.DestroySession(w, r, isHTTPS(r))
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleDashboardProbe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	if !s.auth.Authenticated(r) {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]any{
		"error":   "internal_server_error",
		"message": "Something bad just happened...",
	})
}

func (s *Server) handleAuthenticatedStub(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	if !s.auth.Authenticated(r) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	s.serveIndexWithStatus(w, filepath.Join(s.cfg.Server.StaticDir, "index.html"), http.StatusInternalServerError)
}

func (s *Server) serveStatic(w http.ResponseWriter, r *http.Request, relative string) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	cleanPath := filepath.Clean("/" + relative)
	if strings.Contains(cleanPath, "..") {
		notFound(w)
		return
	}

	if cleanPath != "/" {
		fullPath := filepath.Join(s.cfg.Server.StaticDir, strings.TrimPrefix(cleanPath, "/"))
		if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
			http.ServeFile(w, r, fullPath)
			return
		}
	}

	indexPath := filepath.Join(s.cfg.Server.StaticDir, "index.html")
	s.serveIndexWithStatus(w, indexPath, http.StatusOK)
}

func (s *Server) serveIndexWithStatus(w http.ResponseWriter, indexPath string, status int) {
	data, err := os.ReadFile(indexPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			writePlain(w, http.StatusNotFound, "frontend build is not available")
			return
		}
		writePlain(w, http.StatusInternalServerError, "failed to read frontend build")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(s.renderIndexHTML(data))
}

func (s *Server) renderIndexHTML(data []byte) []byte {
	body := string(data)
	title := html.EscapeString(stripBrandColorTags(s.cfg.Branding.Title))
	description := html.EscapeString(s.cfg.Branding.Description)

	replacements := map[string]string{
		`<title>ProbeShield</title>`:                          `<title>` + title + `</title>`,
		`<meta name="description" content="Authentication" />`: `<meta name="description" content="` + description + `" />`,
	}
	for from, to := range replacements {
		body = strings.ReplaceAll(body, from, to)
	}
	return []byte(body)
}

var brandColorTagPattern = regexp.MustCompile(`\{[0-9a-fA-F]{3,8}\}`)

func stripBrandColorTags(value string) string {
	cleaned := strings.TrimSpace(brandColorTagPattern.ReplaceAllString(value, ""))
	if cleaned == "" {
		return "ProbeShield"
	}
	return cleaned
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func (s *Server) withResponseHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.ResponseHeaders.Enabled {
			headers := s.cfg.ResponseHeaders.Headers
			if len(headers) == 0 {
				headers = defaultResponseHeaders()
			}
			for name, value := range headers {
				w.Header().Set(name, value)
			}
		}
		next.ServeHTTP(w, r)
	})
}

func defaultResponseHeaders() map[string]string {
	return map[string]string{
		"X-Content-Type-Options":       "nosniff",
		"X-XSS-Protection":             "0",
		"X-Frame-Options":              "SAMEORIGIN",
		"Cross-Origin-Opener-Policy":   "same-origin-allow-popups",
		"Cross-Origin-Resource-Policy": "same-site",
		"Referrer-Policy":              "strict-origin-when-cross-origin",
		"Strict-Transport-Security":    "max-age=31536000; includeSubDomains",
		"X-Robots-Tag":                 "noindex, nofollow, noarchive, nosnippet, noimageindex",
		"Content-Security-Policy":      "default-src 'self' *;script-src 'self' 'unsafe-inline' 'unsafe-eval' 'wasm-unsafe-eval' *;img-src 'self' data: *;connect-src 'self' *;worker-src 'self' blob: *;frame-src 'self' *;frame-ancestors 'self' *;base-uri 'self';font-src 'self' https: data:;form-action 'self';object-src 'none';script-src-attr 'none';style-src 'self' https: 'unsafe-inline';upgrade-insecure-requests",
	}
}

func withRequestLog(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		logger.Info("http request", "method", r.Method, "path", r.URL.Path, "status", recorder.status, "duration_ms", time.Since(start).Milliseconds(), "remote", r.RemoteAddr)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(status int) {
	s.status = status
	s.ResponseWriter.WriteHeader(status)
}

func methodNotAllowed(w http.ResponseWriter, allowed string) {
	w.Header().Set("Allow", allowed)
	writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method_not_allowed"})
}

func notFound(w http.ResponseWriter) {
	writeJSON(w, http.StatusNotFound, map[string]any{"error": "not_found"})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writePlain(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(message))
}

func isHTTPS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			if ip := strings.TrimSpace(parts[0]); ip != "" {
				return ip
			}
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
