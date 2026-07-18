package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/proxy"

	"probe-shield/internal/config"
)

type TelegramNotifier struct {
	enabled      bool
	token        string
	notifyTarget string
	proxyURL     string
	brandName    string
	client       *http.Client
}

var brandColorTagPattern = regexp.MustCompile(`\{[0-9a-fA-F]{3,8}\}`)

// stripBrandColorTags mirrors server.renderIndexHTML's stripping logic so the
// Telegram brand label matches what's shown on the auth page.
func stripBrandColorTags(value string) string {
	cleaned := strings.TrimSpace(brandColorTagPattern.ReplaceAllString(value, ""))
	if cleaned == "" {
		return "ProbeShield"
	}
	return cleaned
}

// resolveBrandName picks the Telegram label in priority order:
// dedicated TELEGRAM_BRAND_NAME, then the shared auth-page branding title,
// then the "ProbeShield" default.
func resolveBrandName(cfg config.Config) string {
	if name := strings.TrimSpace(cfg.Telegram.BrandName); name != "" {
		return name
	}
	return stripBrandColorTags(cfg.Branding.Title)
}

func NewTelegramNotifier(cfg config.Config) *TelegramNotifier {
	return &TelegramNotifier{
		enabled:      cfg.Telegram.Enabled,
		token:        cfg.Telegram.BotToken,
		notifyTarget: cfg.Telegram.NotifyTarget,
		proxyURL:     cfg.Telegram.BotProxy,
		brandName:    resolveBrandName(cfg),
		client:       newTelegramHTTPClient(cfg.Telegram.BotProxy),
	}
}

func newTelegramHTTPClient(proxyURL string) *http.Client {
	proxyURL = strings.TrimSpace(proxyURL)
	if proxyURL == "" {
		return &http.Client{Timeout: 10 * time.Second}
	}

	parsed, err := url.Parse(proxyURL)
	if err != nil || parsed.Host == "" {
		return &http.Client{Timeout: 10 * time.Second}
	}

	transport := &http.Transport{}

	switch strings.ToLower(parsed.Scheme) {
	case "socks5", "socks5h":
		var auth *proxy.Auth
		if parsed.User != nil {
			password, _ := parsed.User.Password()
			auth = &proxy.Auth{User: parsed.User.Username(), Password: password}
		}
		dialer, dialErr := proxy.SOCKS5("tcp", parsed.Host, auth, proxy.Direct)
		if dialErr != nil {
			return &http.Client{Timeout: 10 * time.Second}
		}
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		}
	case "http", "https":
		transport.Proxy = http.ProxyURL(parsed)
	default:
		return &http.Client{Timeout: 10 * time.Second}
	}

	return &http.Client{Timeout: 10 * time.Second, Transport: transport}
}

func (t *TelegramNotifier) SendLoginNotification(ctx context.Context, success bool, username, password, clientIP, userAgent, reason string) error {
	if !t.enabled || t.token == "" || t.notifyTarget == "" {
		return nil
	}

	chatID, threadID := parseTelegramTarget(t.notifyTarget)
	if chatID == "" {
		return nil
	}

	var text string
	separator := "➖➖➖➖➖➖➖➖➖"

	if success {
		text = fmt.Sprintf(
			"<tg-emoji emoji-id='5330115548900501467'>🔑</tg-emoji> <tg-emoji emoji-id='5461117441612462242'>✅</tg-emoji> <b>#login_attempt_success</b> (%s)\n%s\n<tg-emoji emoji-id='5256143829672672750'>👥</tg-emoji> <code>%s</code>\n<tg-emoji emoji-id='5447410659077661506'>🌐</tg-emoji> <b>IP:</b> <code>%s</code>\n<tg-emoji emoji-id='5460756166143405924'>💻</tg-emoji> <b>User agent:</b> <code>%s</code>",
			html.EscapeString(t.brandName),
			separator,
			html.EscapeString(username),
			html.EscapeString(clientIP),
			html.EscapeString(userAgent),
		)
	} else {
		text = fmt.Sprintf(
			"<tg-emoji emoji-id='5330115548900501467'>🔑</tg-emoji> <tg-emoji emoji-id='5472267631979405211'>❌</tg-emoji> <b>#login_attempt_failed</b> (%s)\n%s\n<tg-emoji emoji-id='5256143829672672750'>👥</tg-emoji> <code>%s</code>\n<tg-emoji emoji-id='5330115548900501467'>🔑</tg-emoji> <b>Password:</b> <code>%s</code>\n<tg-emoji emoji-id='5447410659077661506'>🌐</tg-emoji> <b>IP:</b> <code>%s</code>\n<tg-emoji emoji-id='5460756166143405924'>💻</tg-emoji> <b>User agent:</b> <code>%s</code>\n<tg-emoji emoji-id='5443038326535759644'>💬</tg-emoji> <b>Description:</b> <code>%s</code>",
			html.EscapeString(t.brandName),
			separator,
			html.EscapeString(username),
			html.EscapeString(password),
			html.EscapeString(clientIP),
			html.EscapeString(userAgent),
			html.EscapeString(reason),
		)
	}

	payload := map[string]any{
		"chat_id":                  chatID,
		"text":                     text,
		"parse_mode":               "HTML",
		"disable_web_page_preview": true,
	}

	if threadID != "" {
		if id, err := strconv.ParseInt(threadID, 10, 64); err == nil {
			payload["message_thread_id"] = id
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	apiURL := "https://api.telegram.org/bot" + strings.TrimSpace(t.token) + "/sendMessage"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("telegram returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	return nil
}

func parseTelegramTarget(rawTarget string) (string, string) {
	target := strings.TrimSpace(rawTarget)
	if target == "" {
		return "", ""
	}
	if !strings.Contains(target, ":") {
		return target, ""
	}
	parts := strings.SplitN(target, ":", 2)
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}
