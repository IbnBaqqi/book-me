package email

import (
	"bytes"
	"context"
	"html/template"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

// TestRealEmailSending (only runs with specific flag)
func TestRealEmailSending(t *testing.T) {
	// Only run when explicitly requested
	if os.Getenv("RUN_EMAIL_TESTS") != "true" {
		t.Skip("Skipping real email test. Set RUN_EMAIL_TESTS=true to run")
	}

	// Load .env file if it exists
	_ = godotenv.Load("../../.env")

	// Check all required env vars upfront
	requiredEnvs := []string{"SMTP_HOST", "SMTP_USERNAME", "SMTP_PASSWORD", "FROM_EMAIL"}
	var missing []string
	for _, env := range requiredEnvs {
		if os.Getenv(env) == "" {
			missing = append(missing, env)
		}
	}
	if len(missing) > 0 {
		t.Skipf("Skipping test: missing required environment variables: %v", missing)
	}

	cfg := Config{
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     587,
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		FromEmail:    os.Getenv("FROM_EMAIL"),
		FromName:     "BookMe Test",
		UseTLS:       true,
	}

	svc, err := NewService(cfg)
	if err != nil {
		t.Fatalf("Failed to create email service: %v", err)
	}

	testEmail := os.Getenv("TEST_RECIPIENT_EMAIL")
	if testEmail == "" {
		testEmail = cfg.SMTPUsername // Send to self
	}

	err = svc.SendConfirmation(
		context.Background(),
		testEmail,
		"Test Conference Room",
		time.Now().Format("Monday, January 2, 2006 at 3:04 PM"),
		time.Now().Add(1*time.Hour).Format("Monday, January 2, 2006 at 3:04 PM"),
	)

	if err != nil {
		t.Errorf("Failed to send email: %v", err)
	} else {
		t.Logf("Email sent successfully to %s", testEmail)
	}
}

// TestTemplateRendering tests that the email template renders correctly
func TestTemplateRendering(t *testing.T) {
	// Only run when explicitly requested
	if os.Getenv("RUN_EMAIL_TESTS") != "true" {
		t.Skip("Skipping template rendering test. Set RUN_EMAIL_TESTS=true to run")
	}

	// Parse the template for the test
	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		t.Fatalf("Failed to parse templates: %v", err)
	}

	// Define mock data
	data := BookingData{
		RoomName:  "Corner",
		StartTime: "Monday, 10:00 AM",
		EndTime:   "Monday, 11:00 AM",
	}

	// Execute the template into a buffer
	var body bytes.Buffer
	err = tmpl.ExecuteTemplate(&body, "confirmation_email_v2.html", data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	// Verify template rendered something
	if body.Len() == 0 {
		t.Fatal("Template rendered empty body")
	}

	// Verify all variables were replaced
	renderedHTML := body.String()
	if strings.Contains(renderedHTML, "{{") {
		t.Error("Template contains unreplaced variables")
	}

	// Verify expected content is present
	expectedContent := []string{
		"Conference",
		"Monday, 10:00 AM",
		"Monday, 11:00 AM",
		"Booking Confirmed",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(renderedHTML, expected) {
			t.Errorf("Template missing expected content: %s", expected)
		}
	}

	// Save to file for manual inspection
	err = os.WriteFile("test_output.html", body.Bytes(), 0600)
	if err != nil {
		t.Logf("Warning: Failed to write debug file: %v", err)
	} else {
		t.Log("Template rendered successfully! Open email/test_output.html in your browser.")
	}
}
