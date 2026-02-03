package google

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Reservation struct {
	StartTime time.Time
	EndTime   time.Time
	CreatedBy string
	Room string
}

type CalendarService struct {
    service    *calendar.Service
    calendarID string
}

// NewCalendarService creates a new calendar service
func NewCalendarService(privateKey, serviceAccountEmail, tokenURI, calendarScope, calendarID string) (*CalendarService, error) {
    ctx := context.Background()

	credentials := map[string]interface{}{
        "type":         "service_account",
        "private_key":  privateKey,
        "client_email": serviceAccountEmail,
        "token_uri":    tokenURI,
    }

    credentialsJSON, err := json.Marshal(credentials)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal credentials: %w", err)
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
        log.Printf("Failed to create calendar event: %v", err)
        return "", fmt.Errorf("failed to create event: %w", err)
    }

    if createdEvent.Id == "" {
        return "", fmt.Errorf("Google Calendar event creation failed: no ID returned")
    }

    return createdEvent.Id, nil
}

// DeleteGoogleEvent deletes a calendar event
func (s *CalendarService) DeleteGoogleEvent(ctx context.Context, eventID string) error {
    err := s.service.Events.Delete(s.calendarID, eventID).Context(ctx).Do()
    if err != nil {
        return fmt.Errorf("failed to delete event: %w", err)
    }

    return nil
}

// func NewCalendarService(tokenManager *TokenManager, calendarURI, calendarID string) *CalendarService {
//     return &CalendarService{
//         client:       &http.Client{Timeout: 30 * time.Second},
//         tokenManager: tokenManager,
//         calendarURI:  calendarURI,
//         calendarID:   calendarID,
//     }
// }


// func (s *CalendarService) CreateGoogleEvent(ctx context.Context, reservation *Reservation) (string, error) {
//     token, err := s.tokenManager.GetAccessToken()
//     if err != nil {
//         log.Printf("Failed to get access token: %v", err)
//         return "", err
//     }

//     location, _ := time.LoadLocation("Europe/Helsinki")
//     start := reservation.StartTime.In(location).Format(time.RFC3339)
//     end := reservation.EndTime.In(location).Format(time.RFC3339)

//     event := EventRequest{
//         Summary:     fmt.Sprintf("[%s] %s meeting room", reservation.CreatedBy, reservation.Room),
//         Description: "Created via BookMe",
//         Start:       DateTimeObject{DateTime: start},
//         End:         DateTimeObject{DateTime: end},
//     }

//     body, err := json.Marshal(event)
//     if err != nil {
//         return "", fmt.Errorf("failed to marshal event: %w", err)
//     }

//     url := fmt.Sprintf("%s/calendars/%s/events", s.calendarURI, s.calendarID)
//     req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
//     if err != nil {
//         return "", fmt.Errorf("failed to create request: %w", err)
//     }

//     req.Header.Set("Authorization", "Bearer "+token)
//     req.Header.Set("Content-Type", "application/json")

//     resp, err := s.client.Do(req)
//     if err != nil {
//         log.Printf("Failed to create calendar event: %v", err)
//         return "", err
//     }
//     defer resp.Body.Close()

//     if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
//         return "", fmt.Errorf("calendar event creation failed with status: %d", resp.StatusCode)
//     }

//     var response Event
//     if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
//         return "", fmt.Errorf("failed to decode response: %w", err)
//     }

//     if response.ID == "" {
//         return "", fmt.Errorf("Google Calendar event creation failed: no ID returned")
//     }

//     return response.ID, nil
// }

// func (s *CalendarService) DeleteGoogleEvent(ctx context.Context, eventID string) error {
//     token, err := s.tokenManager.GetAccessToken()
//     if err != nil {
//         return err
//     }

//     url := fmt.Sprintf("%s/calendars/%s/events/%s", s.calendarURI, s.calendarID, eventID)
//     req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
//     if err != nil {
//         return fmt.Errorf("failed to create request: %w", err)
//     }

//     req.Header.Set("Authorization", "Bearer "+token)

//     resp, err := s.client.Do(req)
//     if err != nil {
//         return err
//     }
//     defer resp.Body.Close()

//     if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
//         return fmt.Errorf("failed to delete event with status: %d", resp.StatusCode)
//     }

//     return nil
// }