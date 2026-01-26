package email

import (
	"bytes"
	"html/template"
	// "os"
	"testing"
)

func TestTemplateRendering(t *testing.T) {
	// Manually parse the template for the test
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

	// Commented to avoid creating it during CI/CD
	// Save to a local HTML file so I can open it in your browser
	// err = os.WriteFile("test_output.html", body.Bytes(), 0644)
	// if err != nil {
	// 	t.Fatalf("Failed to write debug file: %v", err)
	// }

	t.Log("Template rendered successfully! Open email/test_output.html in your browser.")
}