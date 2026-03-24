package calendar

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	calendarEventPath       = "https://www.googleapis.com/calendar/v3/calendars/%s/events"
	calendarEventActionPath = "https://www.googleapis.com/calendar/v3/calendars/%s/events/%s"
)

type EventDate struct {
	Date     string `json:"date"`
	DateTime string `json:"dateTime"`
}

type Event struct {
	ID          string    `json:"id"`
	Summary     string    `json:"summary"`
	Description string    `json:"description"`
	Start       EventDate `json:"start"`
}

type Calendar struct {
	Items []Event `json:"items"`
}

// InsertEvent adds a new calendar event with the provided metadata.
func (c *Client) InsertEvent(ctx context.Context, title, description string, start, end time.Time) error {
	payload := struct {
		Summary     string `json:"summary"`
		Description string `json:"description"`
		Start       struct {
			DateTime string `json:"dateTime"`
		} `json:"start"`
		End struct {
			DateTime string `json:"dateTime"`
		} `json:"end"`
	}{
		Summary:     title,
		Description: description,
	}
	payload.Start.DateTime = start.Format(time.RFC3339)
	payload.End.DateTime = end.Format(time.RFC3339)

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf(calendarEventPath, c.calendarID), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build calendar request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("post event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("calendar error %s: %s", resp.Status, string(respBody))
	}

	return nil
}

// ListEvents retrieves calendar events starting from the first day of the current week.
func (c *Client) ListEvents(ctx context.Context) ([]Event, error) {
	firstDayOfWeek := time.Now().AddDate(0, 0, -int(time.Now().Weekday()))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(calendarEventPath, c.calendarID), nil)
	if err != nil {
		return nil, fmt.Errorf("build calendar request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	q := req.URL.Query()
	q.Add("orderBy", "startTime")
	q.Add("timeMin", firstDayOfWeek.Format(time.RFC3339))
	q.Add("singleEvents", "true")
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("calendar error %s: %s", resp.Status, string(respBody))
	}

	var data Calendar
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode calendar response: %w", err)
	}

	return data.Items, nil
}

// DeleteEvent removes a calendar event by its ID.
func (c *Client) DeleteEvent(ctx context.Context, eventId string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf(calendarEventActionPath, c.calendarID, eventId), nil)
	if err != nil {
		return fmt.Errorf("build calendar request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("calendar error %s: %s", resp.Status, string(respBody))
	}

	return nil
}
