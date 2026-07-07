package config

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	envConfigFile      = "PROBE_SHIELD_CONFIG_FILE"
	envHTTPAddress     = "PROBE_SHIELD_HTTP_ADDRESS"
	envHTTPPort        = "PROBE_SHIELD_HTTP_PORT"
	envStaticDir       = "PROBE_SHIELD_STATIC_DIR"
	envLogin           = "PROBE_SHIELD_LOGIN"
	envPassword        = "PROBE_SHIELD_PASSWORD"
	envSessionTTL      = "PROBE_SHIELD_SESSION_TTL"
	envAuthCheckDelay  = "PROBE_SHIELD_AUTH_CHECK_DELAY"
	envRateWindow      = "PROBE_SHIELD_RL_WINDOW"
	envRateMaxFailures = "PROBE_SHIELD_RL_MAX_FAILURES"
	envBrandTitle      = "PROBE_SHIELD_BRAND_TITLE"
	envBrandLogoURL    = "PROBE_SHIELD_BRAND_LOGO_URL"
	envPageTitle       = "PROBE_SHIELD_PAGE_TITLE"
	envPageDescription = "PROBE_SHIELD_PAGE_DESCRIPTION"
	envHeadersEnabled  = "PROBE_SHIELD_RESPONSE_HEADERS_ENABLED"
	envHeadersJSON     = "PROBE_SHIELD_RESPONSE_HEADERS_JSON"
)

type Config struct {
	Server          ServerConfig          `json:"server"`
	Auth            AuthConfig            `json:"auth"`
	RateLimit       RateLimitConfig       `json:"rateLimit"`
	Branding        BrandingConfig        `json:"branding"`
	PageMeta        PageMetaConfig        `json:"pageMeta"`
	ResponseHeaders ResponseHeadersConfig `json:"responseHeaders"`
}

type ServerConfig struct {
	Address   string `json:"address"`
	Port      int    `json:"port"`
	StaticDir string `json:"staticDir"`
}

type AuthConfig struct {
	Login             string        `json:"login"`
	Password          string        `json:"password"`
	SessionTTL        time.Duration `json:"-"`
	SessionTTLRaw     string        `json:"sessionTTL"`
	AuthCheckDelay    time.Duration `json:"-"`
	AuthCheckDelayRaw string        `json:"authCheckDelay"`
}

type RateLimitConfig struct {
	WindowRaw         string        `json:"window"`
	Window            time.Duration `json:"-"`
	MaxFailedAttempts int           `json:"maxFailedAttempts"`
}

// PageMetaConfig controls browser-level HTML metadata. Values are injected into
// index.html by the backend at runtime, so they can be changed without rebuilding
// the frontend bundle.
type PageMetaConfig struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// ResponseHeadersConfig controls HTTP response headers emitted by the stub.
// If enabled=true and headers contains at least one item, only these configured
// headers are emitted. If enabled=true and headers is empty, built-in Exodus-like
// baseline security headers are emitted. If enabled=false, no security headers are added.
type ResponseHeadersConfig struct {
	Enabled bool              `json:"enabled"`
	Headers map[string]string `json:"headers"`
}

// BrandingConfig mirrors the original frontend status contract, but the values
// are loaded from configuration instead of database settings.
type BrandingConfig struct {
	// Title supports the original colored text format, for example:
	// {8195a3}Exo{eceddb}dus
	Title string `json:"title"`

	// LogoURL may be an absolute URL or a root-relative/static URL, for example:
	// https://example.com/logo.svg or /logo.svg
	LogoURL string `json:"logoUrl"`
}

func Default() Config {
	return Config{
		Server: ServerConfig{
			Address:   "0.0.0.0",
			Port:      3000,
			StaticDir: "../frontend/dist",
		},
		Auth: AuthConfig{
			Login:             "",
			Password:          "",
			SessionTTLRaw:     "12h",
			SessionTTL:        12 * time.Hour,
			AuthCheckDelayRaw: "5s",
			AuthCheckDelay:    5 * time.Second,
		},
		RateLimit: RateLimitConfig{
			WindowRaw:         "5m",
			Window:            5 * time.Minute,
			MaxFailedAttempts: 10,
		},
		Branding: BrandingConfig{
			Title:   "",
			LogoURL: "",
		},
		PageMeta: PageMetaConfig{
			Title:       "shield-probe",
			Description: "Authentication",
		},
		ResponseHeaders: ResponseHeadersConfig{
			Enabled: true,
			Headers: nil,
		},
	}
}

func Load() (Config, error) {
	cfg := Default()
	configPath := strings.TrimSpace(os.Getenv(envConfigFile))
	if configPath == "" {
		configPath = filepath.Join("configs", "probe-shield.json")
	}

	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("parse %s: %w", configPath, err)
		}
	} else if !os.IsNotExist(err) || os.Getenv(envConfigFile) != "" {
		return cfg, fmt.Errorf("read %s: %w", configPath, err)
	}

	applyEnvOverrides(&cfg)
	if err := normalize(&cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if v := strings.TrimSpace(os.Getenv(envHTTPAddress)); v != "" {
		cfg.Server.Address = v
	}
	if v := strings.TrimSpace(os.Getenv(envHTTPPort)); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = p
		}
	}
	if v := strings.TrimSpace(os.Getenv(envStaticDir)); v != "" {
		cfg.Server.StaticDir = v
	}
	if v := os.Getenv(envLogin); v != "" {
		cfg.Auth.Login = v
	}
	if v := os.Getenv(envPassword); v != "" {
		cfg.Auth.Password = v
	}
	if v := strings.TrimSpace(os.Getenv(envSessionTTL)); v != "" {
		cfg.Auth.SessionTTLRaw = v
	}
	if v := strings.TrimSpace(os.Getenv(envAuthCheckDelay)); v != "" {
		cfg.Auth.AuthCheckDelayRaw = v
	}
	if v := strings.TrimSpace(os.Getenv(envRateWindow)); v != "" {
		cfg.RateLimit.WindowRaw = v
	}
	if v := strings.TrimSpace(os.Getenv(envRateMaxFailures)); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.RateLimit.MaxFailedAttempts = n
		}
	}
	if v := strings.TrimSpace(os.Getenv(envBrandTitle)); v != "" {
		cfg.Branding.Title = v
	}
	if v := strings.TrimSpace(os.Getenv(envBrandLogoURL)); v != "" {
		cfg.Branding.LogoURL = v
	}
	if v := strings.TrimSpace(os.Getenv(envPageTitle)); v != "" {
		cfg.PageMeta.Title = v
	}
	if v := strings.TrimSpace(os.Getenv(envPageDescription)); v != "" {
		cfg.PageMeta.Description = v
	}
	if v := strings.TrimSpace(os.Getenv(envHeadersEnabled)); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.ResponseHeaders.Enabled = b
		}
	}
	if v := strings.TrimSpace(os.Getenv(envHeadersJSON)); v != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(v), &headers); err == nil {
			cfg.ResponseHeaders.Headers = headers
		}
	}
}

func normalize(cfg *Config) error {
	if cfg.Server.Address == "" {
		cfg.Server.Address = "0.0.0.0"
	}
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("server.port must be in 1..65535")
	}
	if strings.TrimSpace(cfg.Server.StaticDir) == "" {
		cfg.Server.StaticDir = "../frontend/dist"
	}

	if cfg.Auth.SessionTTLRaw == "" {
		cfg.Auth.SessionTTLRaw = "12h"
	}
	ttl, err := time.ParseDuration(cfg.Auth.SessionTTLRaw)
	if err != nil || ttl <= 0 {
		return fmt.Errorf("auth.sessionTTL must be a positive Go duration, for example 12h")
	}
	cfg.Auth.SessionTTL = ttl

	if cfg.Auth.AuthCheckDelayRaw == "" {
		cfg.Auth.AuthCheckDelayRaw = "5s"
	}
	authDelay, err := time.ParseDuration(cfg.Auth.AuthCheckDelayRaw)
	if err != nil || authDelay < 0 {
		return fmt.Errorf("auth.authCheckDelay must be a non-negative Go duration, for example 5s")
	}
	cfg.Auth.AuthCheckDelay = authDelay

	if cfg.RateLimit.WindowRaw == "" {
		cfg.RateLimit.WindowRaw = "5m"
	}
	window, err := time.ParseDuration(cfg.RateLimit.WindowRaw)
	if err != nil || window <= 0 {
		return fmt.Errorf("rateLimit.window must be a positive Go duration, for example 5m")
	}
	cfg.RateLimit.Window = window
	if cfg.RateLimit.MaxFailedAttempts < 1 {
		return fmt.Errorf("rateLimit.maxFailedAttempts must be >= 1")
	}

	cfg.Branding.Title = strings.TrimSpace(cfg.Branding.Title)
	cfg.Branding.LogoURL = strings.TrimSpace(cfg.Branding.LogoURL)
	cfg.PageMeta.Title = strings.TrimSpace(cfg.PageMeta.Title)
	if cfg.PageMeta.Title == "" {
		cfg.PageMeta.Title = "shield-probe"
	}
	cfg.PageMeta.Description = strings.TrimSpace(cfg.PageMeta.Description)
	if cfg.PageMeta.Description == "" {
		cfg.PageMeta.Description = "Authentication"
	}
	if cfg.ResponseHeaders.Headers != nil {
		normalized := make(map[string]string, len(cfg.ResponseHeaders.Headers))
		for name, value := range cfg.ResponseHeaders.Headers {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			normalized[name] = strings.TrimSpace(value)
		}
		cfg.ResponseHeaders.Headers = normalized
	}
	return nil
}

func (s ServerConfig) ListenAddress() string {
	return net.JoinHostPort(s.Address, strconv.Itoa(s.Port))
}

func (a AuthConfig) IsConfigured() bool {
	return a.Login != "" && a.Password != ""
}
