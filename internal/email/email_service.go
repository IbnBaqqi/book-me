package email

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
	"time"

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
	// Create mail client with TLS
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

// SendConfirmation sends a booking confirmation email
// This runs asynchronously to not block the HTTP response
func (s *Service) SendConfirmation(ctx context.Context, email, room, startTime, endTime string) error {
	// Run in goroutine for async sending
	go func() {
		if err := s.sendConfirmationSync(email, room, startTime, endTime); err != nil {
			// Log error but don't fail the whole operation
			// TODO look into uber-go/zap for logging
			log.Printf("Failed to send confirmation email: %v\n", err)
		}
	}()
	return nil
}

// sendConfirmationSync is the synchronous version for actual email sending
func (s *Service) sendConfirmationSync(toEmail, room, startTime, endTime string) error {
	// Create new message
	msg := mail.NewMsg()

	// Set sender
	if err := msg.From(fmt.Sprintf("%s <%s>", s.fromName, s.from)); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipient
	if err := msg.To(toEmail); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Set subject
	msg.Subject("Hive / Meeting Room Confirmation")

	// Prepare template data
	data := BookingData{
		RoomName:  room,
		StartTime: startTime,
		EndTime:   endTime,
	}

	// Render HTML template
	var htmlBody bytes.Buffer
	if err := s.templates.ExecuteTemplate(&htmlBody, "confirmation_email_v2.html", data); err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// Set HTML body
	msg.SetBodyString(mail.TypeTextHTML, htmlBody.String())

	// plain text fallback
	// plainText := fmt.Sprintf(
	// 	"Hi, the %s meeting room has been reserved for you from %s to %s.",
	// 	room, startTime, endTime,
	// )
	// msg.AddAlternativeString(mail.TypeTextPlain, plainText)

	// Send email with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.client.DialAndSendWithContext(ctx, msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// Close closes the email client connection
func (s *Service) Close() error {
	// wneessen/go-mail client doesn't need explicit closing
	return nil
}