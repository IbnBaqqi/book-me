package email

import (
	"bytes"
	"html/template"
	"os"
	"strings"
	"testing"
	"time"
)

// TestRealEmailSending - Integration test (only runs with specific flag)
func TestRealEmailSending(t *testing.T) {
	// Only run when explicitly requested
	if os.Getenv("RUN_EMAIL_TESTS") != "true" {
		t.Skip("Skipping real email test. Set RUN_EMAIL_TESTS=true to run")
	}

	cfg := Config{
		SMTPHost:     getEnvOrSkip(t, "SMTP_HOST"),
		SMTPPort:     587,
		SMTPUsername: getEnvOrSkip(t, "SMTP_USERNAME"),
		SMTPPassword: getEnvOrSkip(t, "SMTP_PASSWORD"),
		FromEmail:    getEnvOrSkip(t, "FROM_EMAIL"),
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

	err = svc.sendConfirmationSync(
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

	// Skip in CI/CD
	if os.Getenv("CI") != "" || testing.Short() {
		t.Skip("Skipping template rendering test in CI/CD or short mode")
	}

	// Parse the template for the test
	// We use os.DirFS(".") if the test is in the same folder as the templates
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
		"Corner",
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
	err = os.WriteFile("test_output.html", body.Bytes(), 0644)
	if err != nil {
		t.Logf("Warning: Failed to write debug file: %v", err)
	} else {
		t.Log("Template rendered successfully! Open email/test_output.html in your browser.")
	}
}

// Helper function to get environment variable or skip test
func getEnvOrSkip(t *testing.T, key string) string {
	value := os.Getenv(key)
	if value == "" {
		t.Skipf("Skipping test: %s environment variable not set", key)
	}
	return value
}
