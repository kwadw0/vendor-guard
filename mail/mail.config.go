package mail

import (
	"errors"
	"os"
	"strconv"
)

// Config holds SMTP + sender + frontend link settings.
// Loaded from env in cmd/main.go. Empty Host means mail is disabled
// (dev mode: messages are logged, not sent).
type Config struct {
	Host        string
	Port        int
	Username    string
	Password    string
	From        string
	FromName    string
	FrontendURL string
}

// LoadConfig reads mail settings from environment with Mailpit-friendly defaults.
func LoadConfig() Config {
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if port == 0 {
		port = 587
	}
	frontend := os.Getenv("FRONTEND_URL")
	if frontend == "" {
		frontend = "http://localhost:5173"
	}
	from := os.Getenv("MAIL_FROM")
	if from == "" {
		from = "noreply@preuvio.com"
	}
	fromName := os.Getenv("MAIL_FROM_NAME")
	if fromName == "" {
		fromName = "Preuvio"
	}
	return Config{
		Host:        os.Getenv("SMTP_HOST"),
		Port:        port,
		Username:    os.Getenv("SMTP_USER"),
		Password:    os.Getenv("SMTP_PASS"),
		From:        from,
		FromName:    fromName,
		FrontendURL: frontend,
	}
}

// Validate returns error when config would fail at send time.
// Empty host is allowed (disabled mode) — no error.
func (c Config) Validate() error {
	if c.Host == "" {
		return nil
	}
	if c.Port <= 0 || c.Port > 65535 {
		return errors.New("mail: invalid SMTP_PORT")
	}
	if c.From == "" {
		return errors.New("mail: MAIL_FROM is required when SMTP_HOST is set")
	}
	return nil
}

// Enabled reports whether outbound mail will actually be sent.
func (c Config) Enabled() bool {
	return c.Host != ""
}
