// Package google provides Google Calendar integration.
package google

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Reservation represents the data needed to create a calendar event.
type Reservation struct {
	StartTime time.Time
	EndTime   time.Time
	CreatedBy string
	Room string
}

// CalendarService manages Google Calendar operations.
type CalendarService struct {
    service    *calendar.Service
    calendarID string
}

// NewCalendarService creates a new calendar service
func NewCalendarService(credentialsFile, calendarScope, calendarID string) (*CalendarService, error) {

    ctx := context.Background()

	// Read the entire service account JSON file
	credentialsJSON, err := os.ReadFile(credentialsFile) //nolint:gosec // file path comes from config, not user input
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

    // Create JWT config from credentials
    config, err := google.JWTConfigFromJSON(credentialsJSON, calendarScope)
    if err != nil {
        return nil, fmt.Errorf("failed to create JWT config: %w", err)
    }

    // Create calendar service with authenticated client
    service, err := calendar.NewService(ctx, option.WithHTTPClient(config.Client(ctx)))
    if err != nil {
        return nil, fmt.Errorf("failed to create calendar service: %w", err)
    }

    return &CalendarService{
        service:    service,
        calendarID: calendarID,
    }, nil
}

// CreateGoogleEvent creates a calendar event
func (s *CalendarService) CreateGoogleEvent(ctx context.Context, reservation *Reservation) (string, error) {
    location, err := time.LoadLocation("Europe/Helsinki")
    if err != nil {
        return "", fmt.Errorf("failed to load location: %w", err)
    }

    start := reservation.StartTime.In(location)
    end := reservation.EndTime.In(location)

    event := &calendar.Event{
        Summary:     fmt.Sprintf("[%s] %s meeting room", reservation.CreatedBy, reservation.Room),
        Description: "Created via BookMe",
        Start: &calendar.EventDateTime{
            DateTime: start.Format(time.RFC3339),
            TimeZone: "Europe/Helsinki",
        },
        End: &calendar.EventDateTime{
            DateTime: end.Format(time.RFC3339),
            TimeZone: "Europe/Helsinki",
        },
    }

    // Create the event
    createdEvent, err := s.service.Events.Insert(s.calendarID, event).Context(ctx).Do()
    if err != nil {
        slog.Error("failed to create calendar event", "error", err)
        return "", fmt.Errorf("failed to create event: %w", err)
    }

    if createdEvent.Id == "" {
        return "", fmt.Errorf("google calendar event creation failed: no ID returned")
    }

    return createdEvent.Id, nil
}

// HealthCheck verifies the calendar service is accessible
func (s *CalendarService) HealthCheck(ctx context.Context) error {
	// Simple check: try to get calendar metadata
	_, err := s.service.Calendars.Get(s.calendarID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("calendar API unreachable: %w", err)
	}
	return nil
}

// DeleteGoogleEvent deletes a calendar event
func (s *CalendarService) DeleteGoogleEvent(ctx context.Context, eventID string) error {
    err := s.service.Events.Delete(s.calendarID, eventID).Context(ctx).Do()
    if err != nil {
        return fmt.Errorf("failed to delete event: %w", err)
    }

    return nil
}