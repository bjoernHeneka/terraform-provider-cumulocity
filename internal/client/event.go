package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// Event represents a Cumulocity event.
type Event struct {
	ID           string      `json:"id,omitempty"`
	Type         string      `json:"type,omitempty"`
	Text         string      `json:"text,omitempty"`
	Time         string      `json:"time,omitempty"`
	Source       EventSource `json:"source"`
	CreationTime string      `json:"creationTime,omitempty"`
	LastUpdated  string      `json:"lastUpdated,omitempty"`
	Self         string      `json:"self,omitempty"`
}

// EventSource is the managed object reference embedded in an event.
type EventSource struct {
	ID string `json:"id"`
}

type eventCreateRequest struct {
	Type   string      `json:"type"`
	Text   string      `json:"text"`
	Time   string      `json:"time"`
	Source EventSource `json:"source"`
}

type eventUpdateRequest struct {
	Text string `json:"text,omitempty"`
}

type eventCollectionPage struct {
	Events []Event `json:"events"`
	Next   string  `json:"next,omitempty"`
}

// CreateEvent creates a new event in Cumulocity.
func (c *Client) CreateEvent(ctx context.Context, sourceID, eventType, text, time string) (*Event, error) {
	body := eventCreateRequest{
		Type:   eventType,
		Text:   text,
		Time:   time,
		Source: EventSource{ID: sourceID},
	}
	var result Event
	if _, err := c.doJSON(ctx, http.MethodPost, "/event/events", body, &result); err != nil {
		return nil, fmt.Errorf("creating event: %w", err)
	}
	return &result, nil
}

// GetEvent retrieves an event by ID.
func (c *Client) GetEvent(ctx context.Context, id string) (*Event, error) {
	path := fmt.Sprintf("/event/events/%s", url.PathEscape(id))
	var result Event
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, fmt.Errorf("getting event %q: %w", id, err)
	}
	return &result, nil
}

// UpdateEvent updates the text description of an event.
func (c *Client) UpdateEvent(ctx context.Context, id, text string) (*Event, error) {
	path := fmt.Sprintf("/event/events/%s", url.PathEscape(id))
	body := eventUpdateRequest{Text: text}
	var result Event
	if _, err := c.doJSON(ctx, http.MethodPut, path, body, &result); err != nil {
		return nil, fmt.Errorf("updating event %q: %w", id, err)
	}
	return &result, nil
}

// DeleteEvent deletes an event by ID. Returns nil if not found.
func (c *Client) DeleteEvent(ctx context.Context, id string) error {
	path := fmt.Sprintf("/event/events/%s", url.PathEscape(id))
	_, err := c.doJSON(ctx, http.MethodDelete, path, nil, nil)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("deleting event %q: %w", id, err)
	}
	return nil
}

// ListEvents retrieves events with optional filters. All pages are followed.
func (c *Client) ListEvents(ctx context.Context, sourceID, eventType, dateFrom, dateTo string) ([]Event, error) {
	base := "/event/events?pageSize=100"
	if sourceID != "" {
		base += "&source=" + url.QueryEscape(sourceID)
	}
	if eventType != "" {
		base += "&type=" + url.QueryEscape(eventType)
	}
	if dateFrom != "" {
		base += "&dateFrom=" + url.QueryEscape(dateFrom)
	}
	if dateTo != "" {
		base += "&dateTo=" + url.QueryEscape(dateTo)
	}

	var all []Event
	path := base
	for path != "" {
		var page eventCollectionPage
		if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &page); err != nil {
			return nil, fmt.Errorf("listing events: %w", err)
		}
		all = append(all, page.Events...)
		path = ""
		if page.Next != "" {
			u, err := url.Parse(page.Next)
			if err == nil {
				path = u.RequestURI()
			}
		}
	}
	return all, nil
}
