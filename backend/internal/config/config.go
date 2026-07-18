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
	envHTTPAddress      = "PROBE_SHIELD_HTTP_ADDRESS"
	envHTTPPort         = "PROBE_SHIELD_HTTP_PORT"
	envLogin            = "PROBE_SHIELD_LOGIN"
	envPassword         = "PROBE_SHIELD_PASSWORD"
	envRateWindow       = "PROBE_SHIELD_RL_WINDOW"
	envRateMaxFailures  = "PROBE_SHIELD_RL_MAX_FAILURES"
	envBrandTitle       = "PROBE_SHIELD_BRAND_TITLE"
	envBrandDescription = "PROBE_SHIELD_BRAND_DESCRIPTION"
	envBrandLogoURL     = "PROBE_SHIELD_BRAND_LOGO_URL"
	envBrandImageURL    = "PROBE_SHIELD_BRAND_IMAGE_URL"
	envHeadersEnabled   = "PROBE_SHIELD_RESPONSE_HEADERS_ENABLED"
	envHeadersJSON      = "PROBE_SHIELD_RESPONSE_HEADERS_JSON"

	DefaultSessionTTL     = time.Hour
	DefaultAuthCheckDelay = 5 * time.Second
)

type Config struct {
	Server          ServerConfig          `json:"server"`
	Auth            AuthConfig            `json:"auth"`
	RateLimit       RateLimitConfig       `json:"rateLimit"`
	Branding        BrandingConfig        `json:"branding"`
	ResponseHeaders ResponseHeadersConfig `json:"responseHeaders"`
	Telegram        TelegramConfig        `json:"telegram"`
}

type TelegramConfig struct {
	Enabled      bool   `json:"enabled"`
	BotToken     string `json:"botToken"`
	BotProxy     string `json:"botProxy"`
	NotifyTarget string `json:"notifyTarget"`
}

type ServerConfig struct {
	Address   string `json:"address"`
	Port      int    `json:"port"`
	StaticDir string `json:"-"`
}

type AuthConfig struct {
	Login          string        `json:"login"`
	Password       string        `json:"password"`
	SessionTTL     time.Duration `json:"-"`
	AuthCheckDelay time.Duration `json:"-"`
}

type RateLimitConfig struct {
	WindowRaw         string        `json:"window"`
	Window            time.Duration `json:"-"`
	MaxFailedAttempts int           `json:"maxFailedAttempts"`
}

// ResponseHeadersConfig controls HTTP response headers emitted by the stub.
// If enabled=true and headers contains at least one item, only these configured
// headers are emitted. If enabled=true and headers is empty, built-in baseline
// security headers are emitted. If enabled=false, no security headers are added.
type ResponseHeadersConfig struct {
	Enabled bool              `json:"enabled"`
	Headers map[string]string `json:"headers"`
}

// BrandingConfig mirrors the original frontend status contract, but the values
// are loaded from configuration instead of database settings.
type BrandingConfig struct {
	// Title is the single source for both the visible auth-page name and the
	// browser document title. It supports the original colored text format:
	// {8195a3}Exo{eceddb}dus
	Title string `json:"title"`

	// Description is used for the browser meta description.
	Description string `json:"description"`

	// LogoURL may be an absolute URL or a root-relative/static URL, for example:
	// https://example.com/logo.svg or /logo.svg
	LogoURL string `json:"logoUrl"`

	// ImageURL is the preview image used for social media previews (OG, Twitter)
	ImageURL string `json:"imageUrl"`
}

func Default() Config {
	return Config{
		Server: ServerConfig{
			Address:   "0.0.0.0",
			Port:      4000,
			StaticDir: "",
		},
		Auth: AuthConfig{
			Login:          "",
			Password:       "",
			SessionTTL:     DefaultSessionTTL,
			AuthCheckDelay: DefaultAuthCheckDelay,
		},
		RateLimit: RateLimitConfig{
			WindowRaw:         "5m",
			Window:            5 * time.Minute,
			MaxFailedAttempts: 10,
		},
		Branding: BrandingConfig{
			Title:       "",
			Description: "Authentication",
			LogoURL:     "",
			ImageURL:    "",
		},
		ResponseHeaders: ResponseHeadersConfig{
			Enabled: true,
			Headers: nil,
		},
	}
}

func Load() (Config, error) {
	cfg := Default()
	configPath, found, err := findConfigFile()
	if err != nil {
		return cfg, err
	}
	if found {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return cfg, fmt.Errorf("read %s: %w", configPath, err)
		}
		if err := json.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("parse %s: %w", configPath, err)
		}
	}

	applyEnvOverrides(&cfg)
	if err := normalize(&cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func findConfigFile() (path string, found bool, err error) {
	candidates := []string{
		filepath.Join("configs", "config.json"),
		filepath.Join("..", "configs", "config.json"),
		filepath.Join("/app", "configs", "config.json"),
	}
	for _, candidate := range candidates {
		if _, statErr := os.Stat(candidate); statErr == nil {
			return candidate, true, nil
		} else if !os.IsNotExist(statErr) {
			return "", false, fmt.Errorf("stat %s: %w", candidate, statErr)
		}
	}
	return "", false, nil
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
	if v := os.Getenv(envLogin); v != "" {
		cfg.Auth.Login = v
	}
	if v := os.Getenv(envPassword); v != "" {
		cfg.Auth.Password = v
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
	if v := strings.TrimSpace(os.Getenv(envBrandDescription)); v != "" {
		cfg.Branding.Description = v
	}
	if v := strings.TrimSpace(os.Getenv(envBrandLogoURL)); v != "" {
		cfg.Branding.LogoURL = v
	}
	if v := strings.TrimSpace(os.Getenv(envBrandImageURL)); v != "" {
		cfg.Branding.ImageURL = v
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
	if v := strings.TrimSpace(os.Getenv("IS_TELEGRAM_NOTIFICATIONS_ENABLED")); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.Telegram.Enabled = b
		}
	}
	if v := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN")); v != "" {
		cfg.Telegram.BotToken = v
	}
	if v := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_PROXY")); v != "" {
		cfg.Telegram.BotProxy = v
	}
	if v := strings.TrimSpace(os.Getenv("TELEGRAM_NOTIFY_SERVICE")); v != "" {
		cfg.Telegram.NotifyTarget = v
	}
}

func normalize(cfg *Config) error {
	if cfg.Server.Address == "" {
		cfg.Server.Address = "0.0.0.0"
	}
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("server.port must be in 1..65535")
	}
	cfg.Server.StaticDir = detectStaticDir()

	cfg.Auth.SessionTTL = DefaultSessionTTL
	cfg.Auth.AuthCheckDelay = DefaultAuthCheckDelay

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
	cfg.Branding.Description = strings.TrimSpace(cfg.Branding.Description)
	cfg.Branding.LogoURL = strings.TrimSpace(cfg.Branding.LogoURL)
	cfg.Branding.ImageURL = strings.TrimSpace(cfg.Branding.ImageURL)
	if cfg.Branding.ImageURL == "" {
		cfg.Branding.ImageURL = "/favicons/og-image.jpg"
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

func detectStaticDir() string {
	candidates := []string{
		filepath.Join("/app", "frontend", "dist"),
		filepath.Join("..", "frontend", "dist"),
		filepath.Join("frontend", "dist"),
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(filepath.Join(candidate, "index.html")); err == nil && !info.IsDir() {
			return candidate
		}
	}
	return filepath.Join("..", "frontend", "dist")
}

func (s ServerConfig) ListenAddress() string {
	return net.JoinHostPort(s.Address, strconv.Itoa(s.Port))
}

func (a AuthConfig) IsConfigured() bool {
	return a.Login != "" && a.Password != ""
}
