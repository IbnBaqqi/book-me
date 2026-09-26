package email

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type fakeSender struct {
	last  *mail.SGMailV3
	code  int
	err   error
	calls int
}

func (f *fakeSender) SendWithContext(_ context.Context, email *mail.SGMailV3) (*rest.Response, error) {
	f.calls++
	f.last = email
	if f.err != nil {
		return nil, f.err
	}
	if f.code == 0 {
		f.code = http.StatusAccepted
	}
	return &rest.Response{StatusCode: f.code}, nil
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func readyCfg() Config {
	return Config{
		APIKey:                      "sg-test",
		TemplateBookingConfirmation: "d-confirm",
		Production:                  true,
	}
}

func TestSendConfirmationMapsTemplateAndData(t *testing.T) {
	fake := &fakeSender{code: http.StatusAccepted}
	s := newService(testLogger(), readyCfg(), fake)

	if err := s.SendConfirmation(context.Background(), "ada@hive.fi", "Aurora", "09:00", "10:00"); err != nil {
		t.Fatalf("SendConfirmation() error = %v", err)
	}
	if fake.calls != 1 {
		t.Fatalf("calls = %d, want 1", fake.calls)
	}
	if fake.last.TemplateID != "d-confirm" {
		t.Errorf("template = %q, want d-confirm", fake.last.TemplateID)
	}
	if got := fake.last.From.Address; got != fromAddress {
		t.Errorf("from = %q, want %q", got, fromAddress)
	}
	p := fake.last.Personalizations[0]
	if got := p.To[0].Address; got != "ada@hive.fi" {
		t.Errorf("to = %q, want ada@hive.fi", got)
	}
	if got := p.DynamicTemplateData["room_name"]; got != "Aurora" {
		t.Errorf("room_name = %v, want Aurora", got)
	}
	if got := p.DynamicTemplateData["start_time"]; got != "09:00" {
		t.Errorf("start_time = %v, want 09:00", got)
	}
	if got := p.DynamicTemplateData["end_time"]; got != "10:00" {
		t.Errorf("end_time = %v, want 10:00", got)
	}
}

func TestSendConfirmationMissingTemplateIDIsUpstream(t *testing.T) {
	fake := &fakeSender{}
	cfg := readyCfg()
	cfg.TemplateBookingConfirmation = ""
	s := newService(testLogger(), cfg, fake)
	err := s.SendConfirmation(context.Background(), "ada@hive.fi", "Aurora", "09:00", "10:00")
	if !errors.Is(err, ErrUpstream) {
		t.Errorf("error = %v, want ErrUpstream", err)
	}
	if fake.calls != 0 {
		t.Errorf("sender called %d times, want 0", fake.calls)
	}
}

func TestSendStatusMapping(t *testing.T) {
	tests := []struct {
		name string
		code int
		want error
	}{
		{"accepted", http.StatusAccepted, nil},
		{"ok", http.StatusOK, nil},
		{"unauthorized", http.StatusUnauthorized, ErrUnauthorized},
		{"forbidden", http.StatusForbidden, ErrUnauthorized},
		{"bad request", http.StatusBadRequest, ErrUpstream},
		{"server error", http.StatusInternalServerError, ErrUpstream},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeSender{code: tt.code}
			s := newService(testLogger(), readyCfg(), fake)
			err := s.SendConfirmation(context.Background(), "ada@hive.fi", "Aurora", "09:00", "10:00")
			if tt.want == nil {
				if err != nil {
					t.Fatalf("error = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tt.want) {
				t.Errorf("error = %v, want errors.Is(..., %v)", err, tt.want)
			}
		})
	}
}

func TestSendTransportErrorIsUpstream(t *testing.T) {
	fake := &fakeSender{err: errors.New("connection refused")}
	s := newService(testLogger(), readyCfg(), fake)
	err := s.SendConfirmation(context.Background(), "ada@hive.fi", "Aurora", "09:00", "10:00")
	if !errors.Is(err, ErrUpstream) {
		t.Errorf("error = %v, want ErrUpstream", err)
	}
}

func TestEmptyAPIKeyIsNoop(t *testing.T) {
	s := newService(testLogger(), Config{Production: true}, nil)
	if err := s.SendConfirmation(context.Background(), "ada@hive.fi", "Aurora", "09:00", "10:00"); err != nil {
		t.Fatalf("SendConfirmation() error = %v, want nil", err)
	}
}

func TestNonProductionSkipsSend(t *testing.T) {
	fake := &fakeSender{}
	cfg := readyCfg()
	cfg.Production = false
	s := newService(testLogger(), cfg, fake)
	if err := s.SendConfirmation(context.Background(), "ada@hive.fi", "Aurora", "09:00", "10:00"); err != nil {
		t.Fatalf("SendConfirmation() error = %v, want nil", err)
	}
	if fake.calls != 0 {
		t.Errorf("sender called %d times, want 0 (non-production gate)", fake.calls)
	}
}

func TestForceSendOverridesNonProductionGate(t *testing.T) {
	fake := &fakeSender{code: http.StatusAccepted}
	cfg := readyCfg()
	cfg.Production = false
	cfg.ForceSend = true
	s := newService(testLogger(), cfg, fake)
	if err := s.SendConfirmation(context.Background(), "ada@hive.fi", "Aurora", "09:00", "10:00"); err != nil {
		t.Fatalf("SendConfirmation() error = %v, want nil", err)
	}
	if fake.calls != 1 {
		t.Errorf("calls = %d, want 1", fake.calls)
	}
}

func TestEmptyRecipientSkipsSend(t *testing.T) {
	fake := &fakeSender{}
	s := newService(testLogger(), readyCfg(), fake)
	if err := s.SendConfirmation(context.Background(), "  ", "Aurora", "09:00", "10:00"); err != nil {
		t.Fatalf("SendConfirmation() error = %v, want nil", err)
	}
	if fake.calls != 0 {
		t.Errorf("sender called %d times, want 0", fake.calls)
	}
}
