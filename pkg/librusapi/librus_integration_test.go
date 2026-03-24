package librusapi

import (
	"context"
	"librendarium/pkg/config"
	"testing"
	"time"
)

func TestIntegrationLibrus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	cfg, err := config.Load("../../.env")
	if err != nil {
		t.Fatalf("failed to load config: %v (make sure .env is present with real credentials)", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := NewClient(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to create librus client: %v", err)
	}

	hw, err := client.GetHomeWorks(ctx)
	if err != nil {
		t.Fatalf("failed to get homeworks: %v", err)
	}

	t.Logf("Successfully fetched %d homeworks", len(hw.HomeWorks))

	if len(hw.HomeWorks) > 0 {
		first := hw.HomeWorks[0]
		if first.Subject != nil {
			subj, err := client.GetSubject(ctx, first.Subject.ID)
			if err != nil {
				t.Errorf("failed to get subject %d: %v", first.Subject.ID, err)
			} else {
				t.Logf("Subject detail: %s", subj.Subject.Name)
			}
		}

		cat, err := client.GetCategory(ctx, first.Category.ID)
		if err != nil {
			t.Errorf("failed to get category %d: %v", first.Category.ID, err)
		} else {
			t.Logf("Category detail: %s", cat.Category.Name)
		}
	}
}
