package mail

import (
	"context"
	"log/slog"
	"strings"
	"time"

	gomail "github.com/wneessen/go-mail"
)

// Message is a single outbound email.
type Message struct {
	To      []string
	Subject string
	HTML    string
	Text    string
}

// Mailer is the interface consumed by auth/partners services.
// Implemented by *Service; allows mocks in tests.
type Mailer interface {
	SendVerificationEmail(ctx context.Context, to, name, token string) error
	SendPartnerInviteEmail(ctx context.Context, to, partnerName, token string) error
	SendPasswordResetEmail(ctx context.Context, to, name, token string) error
	Send(ctx context.Context, msg *Message) error
}

// Service sends mail via SMTP (wneessen/go-mail) with a bounded
// in-process queue so HTTP handlers never block on SMTP.
type Service struct {
	cfg    Config
	logger *slog.Logger
	queue  chan queuedMail
}

type queuedMail struct {
	msg *Message
}

// NewService builds a mail service and starts 2 background workers.
// If cfg.Host is empty, mail is disabled: Send logs instead of dialing.
func NewService(cfg Config, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	s := &Service{cfg: cfg, logger: logger, queue: make(chan queuedMail, 256)}
	for i := 0; i < 2; i++ {
		go s.worker()
	}
	return s
}

func (s *Service) worker() {
	for q := range s.queue {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		if err := s.sendNow(ctx, q.msg); err != nil {
			s.logger.Error("mail send failed", "to", q.msg.To, "subject", q.msg.Subject, "error", err)
		}
		cancel()
	}
}

// Send enqueues for async delivery (non-blocking) and also attempts
// synchronous send when queue is full to avoid silent drops.
func (s *Service) Send(ctx context.Context, msg *Message) error {
	select {
	case s.queue <- queuedMail{msg: msg}:
		return nil
	default:
		return s.sendNow(ctx, msg)
	}
}

// SendSync bypasses the queue (used by workers and tests).
func (s *Service) SendSync(ctx context.Context, msg *Message) error {
	return s.sendNow(ctx, msg)
}

func (s *Service) sendNow(ctx context.Context, msg *Message) error {
	if !s.cfg.Enabled() {
		s.logger.Info("mail disabled (no SMTP_HOST), logging instead",
			"to", msg.To, "subject", msg.Subject)
		return nil
	}
	m := gomail.NewMsg()
	// go-mail signature is FromFormat(name, addr) — name first.
	if err := m.FromFormat(s.cfg.FromName, s.cfg.From); err != nil {
		return err
	}
	if err := m.To(msg.To...); err != nil {
		return err
	}
	m.Subject(msg.Subject)
	if msg.Text != "" {
		m.SetBodyString(gomail.TypeTextPlain, msg.Text)
	}
	if msg.HTML != "" {
		if msg.Text != "" {
			m.AddAlternativeString(gomail.TypeTextHTML, msg.HTML)
		} else {
			m.SetBodyString(gomail.TypeTextHTML, msg.HTML)
		}
	}

	client, err := gomail.NewClient(
		s.cfg.Host,
		gomail.WithPort(s.cfg.Port),
		gomail.WithSMTPAuth(gomail.SMTPAuthPlain),
		gomail.WithUsername(s.cfg.Username),
		gomail.WithPassword(s.cfg.Password),
		gomail.WithTimeout(15*time.Second),
	)
	if err != nil {
		return err
	}
	return client.DialAndSendWithContext(ctx, m)
}

// SendVerificationEmail queues the signup verification mail.
// Link points at FRONTEND_URL/verify-email?token= so the SPA can call the API.
func (s *Service) SendVerificationEmail(ctx context.Context, to, name, token string) error {
	link := strings.TrimRight(s.cfg.FrontendURL, "/") + "/verify-email?token=" + token
	subject, html, text := renderVerifyEmail(name, link)
	return s.Send(ctx, &Message{To: []string{to}, Subject: subject, HTML: html, Text: text})
}

// SendPartnerInviteEmail queues the partner-user invitation mail.
func (s *Service) SendPartnerInviteEmail(ctx context.Context, to, partnerName, token string) error {
	link := strings.TrimRight(s.cfg.FrontendURL, "/") + "/invite/accept?token=" + token
	subject, html, text := renderPartnerInvite(partnerName, link)
	return s.Send(ctx, &Message{To: []string{to}, Subject: subject, HTML: html, Text: text})
}

// SendPasswordResetEmail queues the password-reset mail.
// The reset endpoint is not wired yet; this function is ready for it.
func (s *Service) SendPasswordResetEmail(ctx context.Context, to, name, token string) error {
	link := strings.TrimRight(s.cfg.FrontendURL, "/") + "/reset-password?token=" + token
	subject, html, text := renderPasswordReset(name, link)
	return s.Send(ctx, &Message{To: []string{to}, Subject: subject, HTML: html, Text: text})
}
