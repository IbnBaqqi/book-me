// Package email is the SendGrid-backed transactional email service.
// Callers use the named method SendConfirmation; template ids and the
// SendGrid JSON schema stay here. cmd constructs the Service; consuming
// domains depend on an interface they own.
package email

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// Sentinels returned by Service methods. Use errors.Is to match. They never
// drive user-facing copy.
var (
	// ErrUnauthorized means SendGrid returned 401 or 403: the API key is
	// wrong or lacks mail.send. Log loudly.
	ErrUnauthorized = errors.New("email: unauthorized")
	// ErrUpstream wraps every other failure (other 4xx, 5xx, transport).
	ErrUpstream = errors.New("email: upstream error")
)

const (
	fromAddress = "noreply@hive.fi"
	sendTimeout = 5 * time.Second

	kindBookingConfirmation = "booking_confirmation"
)

// sender is the sendgrid.Client method we call. Tests substitute a fake.
type sender interface {
	SendWithContext(ctx context.Context, email *mail.SGMailV3) (*rest.Response, error)
}

// Config is the constructor input. Empty APIKey disables sending (methods
// become no-ops). Outside production the service logs the would-be email and
// does not send unless ForceSend.
type Config struct {
	APIKey                      string
	TemplateBookingConfirmation string
	ForceSend                   bool
	Production                  bool
}

// Service sends named transactional emails through SendGrid dynamic templates.
type Service struct {
	logger                      *slog.Logger
	sender                      sender
	from                        *mail.Email
	templateBookingConfirmation string
	production                  bool
	forceSend                   bool
}

// NewService builds a Service. An empty API key leaves sender nil so every
// method is a no-op. In production that also logs a loud startup warning.
func NewService(logger *slog.Logger, cfg Config) *Service {
	return newService(logger, cfg, productionSender(cfg.APIKey))
}

func productionSender(apiKey string) sender {
	if apiKey == "" {
		return nil
	}
	// New client per send: sendgrid.Client mutates Request.Body and is not
	// safe to share across concurrent requests.
	return senderFunc(func(ctx context.Context, email *mail.SGMailV3) (*rest.Response, error) {
		return sendgrid.NewSendClient(apiKey).SendWithContext(ctx, email)
	})
}

type senderFunc func(ctx context.Context, email *mail.SGMailV3) (*rest.Response, error)

func (f senderFunc) SendWithContext(ctx context.Context, email *mail.SGMailV3) (*rest.Response, error) {
	return f(ctx, email)
}

func newService(logger *slog.Logger, cfg Config, s sender) *Service {
	svc := &Service{
		logger:                      logger,
		sender:                      s,
		from:                        mail.NewEmail("", fromAddress),
		templateBookingConfirmation: cfg.TemplateBookingConfirmation,
		production:                  cfg.Production,
		forceSend:                   cfg.ForceSend,
	}
	if cfg.Production && s == nil {
		logger.Warn("SENDGRID_API_KEY is empty; transactional email is dark")
	}
	return svc
}

// SendConfirmation sends a booking confirmation email.
// Dynamic template data: room_name (string), start_time (string), end_time (string).
func (s *Service) SendConfirmation(ctx context.Context, toEmail, room, startTime, endTime string) error {
	return s.send(ctx, kindBookingConfirmation, toEmail, s.templateBookingConfirmation, map[string]any{
		"room_name":  room,
		"start_time": startTime,
		"end_time":   endTime,
	})
}

func (s *Service) send(ctx context.Context, kind, to, templateID string, data map[string]any) error {
	if s.sender == nil {
		return nil
	}
	to = strings.TrimSpace(to)
	if to == "" {
		s.logger.Warn("email skipped (empty recipient)", "kind", kind)
		return nil
	}
	if !s.production && !s.forceSend {
		s.logger.Info("email skipped (non-production)", "kind", kind, "to", to)
		return nil
	}
	if templateID == "" {
		err := fmt.Errorf("%w: missing template id for %s", ErrUpstream, kind)
		s.fail(ctx, kind, err)
		return err
	}

	msg := mail.NewV3Mail()
	msg.SetFrom(s.from)
	msg.SetTemplateID(templateID)
	p := mail.NewPersonalization()
	p.AddTos(mail.NewEmail("", to))
	for k, v := range data {
		p.SetDynamicTemplateData(k, v)
	}
	msg.AddPersonalizations(p)

	ctx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()

	resp, err := s.sender.SendWithContext(ctx, msg)
	if err != nil {
		err = fmt.Errorf("%w: %v", ErrUpstream, err)
		s.fail(ctx, kind, err)
		return err
	}
	if mapped := mapStatus(resp.StatusCode); mapped != nil {
		s.fail(ctx, kind, mapped)
		return mapped
	}
	return nil
}

func (s *Service) fail(ctx context.Context, kind string, err error) {
	if errors.Is(err, ErrUnauthorized) {
		s.logger.Error("sendgrid API key rejected", "kind", kind, "err", err)
	} else {
		s.logger.Error("email send failed", "kind", kind, "err", err)
	}
}

func mapStatus(status int) error {
	if status >= 200 && status < 300 {
		return nil
	}
	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		return fmt.Errorf("%w: status %d", ErrUnauthorized, status)
	}
	return fmt.Errorf("%w: status %d", ErrUpstream, status)
}
