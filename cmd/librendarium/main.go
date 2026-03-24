package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"librendarium/pkg/calendar"
	"librendarium/pkg/config"
	"librendarium/pkg/librusapi"

	"golang.org/x/sync/errgroup"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if cfg.SyncInterval == "" {
		// Run once and exit
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		if err := runSync(ctx, cfg); err != nil {
			log.Fatalf("sync failed: %v", err)
		}
		return
	}

	interval, err := time.ParseDuration(cfg.SyncInterval)
	if err != nil {
		log.Fatalf("invalid SYNC_INTERVAL %q: %v", cfg.SyncInterval, err)
	}

	log.Printf("starting in daemon mode (sync every %s)", interval)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Initial sync
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	if err := runSync(ctx, cfg); err != nil {
		log.Printf("initial sync failed: %v", err)
	}
	cancel()

	for {
		select {
		case <-sigChan:
			log.Println("shutting down...")
			return
		case <-ticker.C:
			log.Println("starting periodic sync...")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			if err := runSync(ctx, cfg); err != nil {
				log.Printf("periodic sync failed: %v", err)
			}
			cancel()
			log.Println("periodic sync finished")
		}
	}
}

func runSync(ctx context.Context, cfg config.Config) error {
	librusClient, err := librusapi.NewClient(ctx, cfg)
	if err != nil {
		return fmt.Errorf("librus auth: %w", err)
	}

	homeworksResp, err := librusClient.GetHomeWorks(ctx)
	if err != nil {
		return fmt.Errorf("fetch homeworks: %w", err)
	}

	if len(homeworksResp.HomeWorks) == 0 {
		log.Println("no homeworks found")
		return nil
	}

	calClient, err := calendar.NewClient(ctx, cfg)
	if err != nil {
		return fmt.Errorf("calendar client: %w", err)
	}

	// Fetch additional data concurrently
	subjectCache, categoryCache, err := fetchMetadata(ctx, librusClient, homeworksResp.HomeWorks)
	if err != nil {
		return fmt.Errorf("fetch metadata: %w", err)
	}

	existingEvents, err := calClient.ListEvents(ctx)
	if err != nil {
		return fmt.Errorf("list events: %w", err)
	}

	return syncEvents(ctx, calClient, homeworksResp.HomeWorks, existingEvents, subjectCache, categoryCache)
}

// fetchMetadata fetches subjects and categories for given homeworks concurrently.
func fetchMetadata(ctx context.Context, client *librusapi.Client, homeworks []librusapi.HomeWork) (map[uint32]string, map[uint32]string, error) {
	subjectIDs := map[uint32]struct{}{}
	categoryIDs := map[uint32]struct{}{}

	for _, hw := range homeworks {
		if hw.Subject != nil {
			subjectIDs[hw.Subject.ID] = struct{}{}
		}
		categoryIDs[hw.Category.ID] = struct{}{}
	}

	subjectNames := make(map[uint32]string)
	categoryNames := make(map[uint32]string)

	g, ctx := errgroup.WithContext(ctx)
	var mu sync.Mutex

	for id := range subjectIDs {
		id := id
		g.Go(func() error {
			resp, err := client.GetSubject(ctx, id)
			if err != nil {
				return err
			}

			mu.Lock()
			subjectNames[id] = resp.Subject.Name
			mu.Unlock()
			return nil
		})
	}

	for id := range categoryIDs {
		id := id
		g.Go(func() error {
			resp, err := client.GetCategory(ctx, id)
			if err != nil {
				return err
			}

			mu.Lock()
			categoryNames[id] = resp.Category.Name
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, nil, err
	}

	return subjectNames, categoryNames, nil
}

// syncEvents synchronizes homeworks with Google Calendar events.
func syncEvents(ctx context.Context, calClient *calendar.Client, homeworks []librusapi.HomeWork, existingEvents []calendar.Event, subjects, categories map[uint32]string) error {
	today := time.Now().Truncate(24 * time.Hour)

	eventIndex := make(map[string]calendar.Event)
	for _, e := range existingEvents {
		start, err := parseEventTime(e.Start)
		if err != nil {
			continue
		}

		key := fmt.Sprintf("%s|%d", e.Summary, start.Unix())
		eventIndex[key] = e
	}

	syncedEventIDs := sync.Map{}

	g, ctx := errgroup.WithContext(ctx)

	for _, hw := range homeworks {
		hw := hw

		g.Go(func() error {

			start, end, err := librusapi.BuildEventTimes(hw)
			if err != nil {
				log.Printf("skipping homework %d: %v", hw.ID, err)
				return nil
			}

			if start.Before(today) {
				return nil
			}

			title := "Homework"

			if hw.Subject != nil {
				if s, ok := subjects[hw.Subject.ID]; ok {
					title = s
				}
			}

			if cat, ok := categories[hw.Category.ID]; ok {
				title = fmt.Sprintf("%s – %s", title, cat)
			}

			key := fmt.Sprintf("%s|%d", title, start.Unix())

			if e, ok := eventIndex[key]; ok {
				syncedEventIDs.Store(e.ID, struct{}{})
				return nil
			}

			if err := calClient.InsertEvent(ctx, title, hw.Content, start, end); err != nil {
				return fmt.Errorf("insert event: %w", err)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	// Delete events that are no longer in Librus and are in the future
	deleteCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	g2, deleteCtx := errgroup.WithContext(deleteCtx)

	for _, event := range existingEvents {
		event := event

		g2.Go(func() error {

			if _, ok := syncedEventIDs.Load(event.ID); ok {
				return nil
			}

			if err := calClient.DeleteEvent(deleteCtx, event.ID); err != nil {
				log.Printf("failed to delete stale event %s: %v", event.ID, err)
			}

			return nil
		})
	}

	return g2.Wait()
}

func parseEventTime(ed calendar.EventDate) (time.Time, error) {
	if ed.DateTime != "" {
		return time.Parse(time.RFC3339, ed.DateTime)
	}
	return time.Parse("2006-01-02", ed.Date)
}
