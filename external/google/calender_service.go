package google

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"
)

type CalendarService struct {
    client       *http.Client
    tokenManager *TokenManager
    calendarURI  string
    calendarID   string
}

func NewCalendarService(tokenManager *TokenManager, calendarURI, calendarID string) *CalendarService {
    return &CalendarService{
        client:       &http.Client{Timeout: 30 * time.Second},
        tokenManager: tokenManager,
        calendarURI:  calendarURI,
        calendarID:   calendarID,
    }
}

// Assuming you have a Reservation struct somewhere
type Reservation struct {
    StartTime time.Time
    EndTime   time.Time
    CreatedBy string
    Room string
}

func (s *CalendarService) CreateGoogleEvent(ctx context.Context, reservation *Reservation) (string, error) {
    token, err := s.tokenManager.GetAccessToken()
    if err != nil {
        log.Printf("Failed to get access token: %v", err)
        return "", err
    }

    location, _ := time.LoadLocation("Europe/Helsinki")
    start := reservation.StartTime.In(location).Format(time.RFC3339)
    end := reservation.EndTime.In(location).Format(time.RFC3339)

    event := EventRequest{
        Summary:     fmt.Sprintf("[%s] %s meeting room", reservation.CreatedBy, reservation.Room),
        Description: "Created via BookMe",
        Start:       DateTimeObject{DateTime: start},
        End:         DateTimeObject{DateTime: end},
    }

    body, err := json.Marshal(event)
    if err != nil {
        return "", fmt.Errorf("failed to marshal event: %w", err)
    }

    url := fmt.Sprintf("%s/calendars/%s/events", s.calendarURI, s.calendarID)
    req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
    if err != nil {
        return "", fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("Content-Type", "application/json")

    resp, err := s.client.Do(req)
    if err != nil {
        log.Printf("Failed to create calendar event: %v", err)
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
        return "", fmt.Errorf("calendar event creation failed with status: %d", resp.StatusCode)
    }

    var response Event
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return "", fmt.Errorf("failed to decode response: %w", err)
    }

    if response.ID == "" {
        return "", fmt.Errorf("Google Calendar event creation failed: no ID returned")
    }

    return response.ID, nil
}

func (s *CalendarService) DeleteGoogleEvent(ctx context.Context, eventID string) error {
    token, err := s.tokenManager.GetAccessToken()
    if err != nil {
        return err
    }

    url := fmt.Sprintf("%s/calendars/%s/events/%s", s.calendarURI, s.calendarID, eventID)
    req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Set("Authorization", "Bearer "+token)

    resp, err := s.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
        return fmt.Errorf("failed to delete event with status: %d", resp.StatusCode)
    }

    return nil
}