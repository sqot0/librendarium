package librusapi

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *Client) GetHomeWorks(ctx context.Context) (ResponseHomeWorks, error) {
	body, err := c.getAPI(ctx, "HomeWorks")
	if err != nil {
		return ResponseHomeWorks{}, err
	}

	var payload ResponseHomeWorks
	if err := json.Unmarshal(body, &payload); err != nil {
		return ResponseHomeWorks{}, fmt.Errorf("unmarshal homeworks: %w", err)
	}
	return payload, nil
}

func (c *Client) GetCategory(ctx context.Context, id uint32) (ResponseCategory, error) {
	body, err := c.getAPI(ctx, fmt.Sprintf("HomeWorks/Categories/%d", id))
	if err != nil {
		return ResponseCategory{}, err
	}
	var payload ResponseCategory
	if err := json.Unmarshal(body, &payload); err != nil {
		return ResponseCategory{}, fmt.Errorf("unmarshal category: %w", err)
	}
	return payload, nil
}

func (c *Client) GetSubject(ctx context.Context, id uint32) (ResponseSubject, error) {
	body, err := c.getAPI(ctx, fmt.Sprintf("Subjects/%d", id))
	if err != nil {
		return ResponseSubject{}, err
	}
	var payload ResponseSubject
	if err := json.Unmarshal(body, &payload); err != nil {
		return ResponseSubject{}, fmt.Errorf("unmarshal subject: %w", err)
	}
	return payload, nil
}
