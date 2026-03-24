package calendar

import (
	"context"
	"librendarium/pkg/config"
	"testing"
	"time"
)

func TestIntegrationCalendar(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	cfg, err := config.Load("../../.env")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := NewClient(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to create calendar client: %v", err)
	}

	events, err := client.ListEvents(ctx)
	if err != nil {
		t.Fatalf("failed to list events: %v", err)
	}

	t.Logf("Successfully fetched %d events from calendar %s", len(events), cfg.GoogleCalendarID)

	// Test event insertion and deletion as a lifecycle test if we have access.
	// Use a dummy title with current timestamp to be unique.
	testTitle := "Test Integration Event " + time.Now().Format(time.RFC3339)
	start := time.Now().Add(1 * time.Hour)
	end := start.Add(1 * time.Hour)

	err = client.InsertEvent(ctx, testTitle, "This is a test event from Integration test", start, end)
	if err != nil {
		t.Fatalf("failed to insert test event: %v", err)
	}
	t.Log("Successfully inserted test event")

	// Verify insertion and cleanup
	eventsAfter, err := client.ListEvents(ctx)
	if err != nil {
		t.Fatalf("failed to list events after insertion: %v", err)
	}

	var createdID string
	for _, e := range eventsAfter {
		if e.Summary == testTitle {
			createdID = e.ID
			break
		}
	}

	if createdID == "" {
		t.Error("test event was not found in listing after creation")
	} else {
		err = client.DeleteEvent(ctx, createdID)
		if err != nil {
			t.Errorf("failed to delete test event %s: %v", createdID, err)
		} else {
			t.Log("Successfully deleted test event")
		}
	}
}
