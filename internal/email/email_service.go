// Package email provides email notification services.
package email

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"time"

	"github.com/avast/retry-go/v5"
	"github.com/wneessen/go-mail"
)

//go:embed templates/*.html
var templateFS embed.FS

// Service handles email operations
type Service struct {
	client   *mail.Client
	from     string
	fromName string
	templates *template.Template
}

// Config holds email service configuration
type Config struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
	UseTLS       bool
}

// BookingData holds data for booking confirmation email
type BookingData struct {
	RoomName  string
	StartTime string
	EndTime   string
}

// NewService creates a new email service
func NewService(cfg Config) (*Service, error) {
	
	client, err := mail.NewClient(
		cfg.SMTPHost,
		mail.WithPort(cfg.SMTPPort),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(cfg.SMTPUsername),
		mail.WithPassword(cfg.SMTPPassword),
		mail.WithTLSPolicy(mail.TLSMandatory),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create mail client: %w", err)
	}

	// Parse all templates
	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse email templates: %w", err)
	}

	return &Service{
		client:    client,
		from:      cfg.FromEmail,
		fromName:  cfg.FromName,
		templates: tmpl,
	}, nil
}


// SendConfirmation sends a confirmation email for reservation
func (s *Service) SendConfirmation(ctx context.Context, toEmail, room, startTime, endTime string) error {

	msg := mail.NewMsg()

	if err := msg.From(fmt.Sprintf("%s <%s>", s.fromName, s.from)); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	if err := msg.To(toEmail); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	msg.Subject("Hive / Meeting Room Confirmation")

	data := BookingData{
		RoomName:  room,
		StartTime: startTime,
		EndTime:   endTime,
	}

	var htmlBody bytes.Buffer
	if err := s.templates.ExecuteTemplate(&htmlBody, "confirmation_email_v2.html", data); err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	msg.SetBodyString(mail.TypeTextHTML, htmlBody.String())

	// plain text fallback
	// plainText := fmt.Sprintf(
	// 	"Hi, the %s meeting room has been reserved for you from %s to %s.",
	// 	room, startTime, endTime,
	// )
	// msg.AddAlternativeString(mail.TypeTextPlain, plainText)

	// Send email with context and backoff retries
	return retry.Do(
		func() error {
			return s.client.DialAndSendWithContext(ctx, msg)
		},
		retry.Attempts(3),
		retry.Delay(4*time.Second),
		retry.MaxDelay(10*time.Second),
		retry.DelayType(retry.BackOffDelay),
		retry.Context(ctx),
		retry.OnRetry(func(n uint, err error) {
			slog.Warn("retrying email send", 
				"attempt", n+1, 
				"to", toEmail,
				"error", err)
		}),
	)
}

// Close closes the email client connection
// wneessen/go-mail client doesn't need explicit closing
func (s *Service) Close() error {
	return nil
}