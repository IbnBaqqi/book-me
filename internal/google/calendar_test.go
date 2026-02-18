package google

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

const testCalendarScope = "https://www.googleapis.com/auth/calendar"

// setupIntegrationTest creates a CalendarService for integration tests.
// It skips the test if required environment variables are not set.
func setupIntegrationTest(t *testing.T) *CalendarService {
	t.Helper()

	if os.Getenv("RUN_CALENDAR_TESTS") != "true" {
		t.Skip("skipping calendar integration test. Set RUN_CALENDAR_TESTS=true to run)")
	}

	_ = godotenv.Load("../../.env")
	
	credFile := os.Getenv("GOOGLE_CREDENTIALS_FILE")
	calID := os.Getenv("GOOGLE_CALENDAR_ID")

	if credFile == "" || calID == "" {
		t.Skip("missing GOOGLE_CREDENTIALS_FILE or GOOGLE_CALENDAR_ID")
	}

	svc, err := NewCalendarService(credFile, testCalendarScope, calID)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	return svc
}

func TestNewCalendarService_Integration(t *testing.T) {
	svc := setupIntegrationTest(t)

	if svc.service == nil {
		t.Error("expected non-nil calendar service")
	}
	if svc.calendarID == "" {
		t.Error("expected non-empty calendarID")
	}
}

func TestNewCalendarService_InvalidCredentialsFile(t *testing.T) {
	_, err := NewCalendarService("nonexistent.json", testCalendarScope, "some-id")
	if err == nil {
		t.Fatal("expected error for missing credentials file")
	}
}

func TestNewCalendarService_InvalidJSON(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "bad-creds-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()
	if _, err := tmpFile.WriteString("not valid json"); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	_ = tmpFile.Close()

	_, err = NewCalendarService(tmpFile.Name(), testCalendarScope, "some-id")
	if err == nil {
		t.Fatal("expected error for invalid JSON credentials")
	}
}

func TestHealthCheck_Integration(t *testing.T) {
	svc := setupIntegrationTest(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := svc.HealthCheck(ctx); err != nil {
		t.Errorf("health check failed: %v", err)
	}
}

func TestCreateAndDeleteEvent_Integration(t *testing.T) {
	svc := setupIntegrationTest(t)

	reservation := &Reservation{
		StartTime: time.Now().Add(24 * time.Hour),
		EndTime:   time.Now().Add(25 * time.Hour),
		CreatedBy: "Test User",
		Room:      "Test Room",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	eventID, err := svc.CreateGoogleEvent(ctx, reservation)
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}
	if eventID == "" {
		t.Fatal("expected non-empty event ID")
	}

	defer func() {
		delCtx, delCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer delCancel()
		if err := svc.DeleteGoogleEvent(delCtx, eventID); err != nil {
			t.Errorf("failed to delete event: %v", err)
		}
	}()
}

func TestDeleteGoogleEvent_InvalidID_Integration(t *testing.T) {
	svc := setupIntegrationTest(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := svc.DeleteGoogleEvent(ctx, "nonexistent-event-id")
	if err == nil {
		t.Error("expected error when deleting nonexistent event")
	}
}
