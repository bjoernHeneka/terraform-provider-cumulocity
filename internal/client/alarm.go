package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// Alarm represents a Cumulocity alarm.
type Alarm struct {
	ID                  string      `json:"id,omitempty"`
	Type                string      `json:"type,omitempty"`
	Text                string      `json:"text,omitempty"`
	Severity            string      `json:"severity,omitempty"`
	Status              string      `json:"status,omitempty"`
	Time                string      `json:"time,omitempty"`
	Source              AlarmSource `json:"source"`
	Count               int64       `json:"count,omitempty"`
	CreationTime        string      `json:"creationTime,omitempty"`
	LastUpdated         string      `json:"lastUpdated,omitempty"`
	FirstOccurrenceTime string      `json:"firstOccurrenceTime,omitempty"`
	Self                string      `json:"self,omitempty"`
}

// AlarmSource is the managed object reference embedded in an alarm.
type AlarmSource struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

type alarmCreateRequest struct {
	Type     string      `json:"type"`
	Text     string      `json:"text"`
	Severity string      `json:"severity"`
	Status   string      `json:"status,omitempty"`
	Time     string      `json:"time"`
	Source   AlarmSource `json:"source"`
}

type alarmUpdateRequest struct {
	Text     string `json:"text,omitempty"`
	Status   string `json:"status,omitempty"`
	Severity string `json:"severity,omitempty"`
}

type alarmCollectionPage struct {
	Alarms []Alarm `json:"alarms"`
	Next   string  `json:"next,omitempty"`
}

// CreateAlarm creates a new alarm in Cumulocity.
func (c *Client) CreateAlarm(ctx context.Context, sourceID, alarmType, text, severity, status, time string) (*Alarm, error) {
	body := alarmCreateRequest{
		Type:     alarmType,
		Text:     text,
		Severity: severity,
		Status:   status,
		Time:     time,
		Source:   AlarmSource{ID: sourceID},
	}
	var result Alarm
	if _, err := c.doJSON(ctx, http.MethodPost, "/alarm/alarms", body, &result); err != nil {
		return nil, fmt.Errorf("creating alarm: %w", err)
	}
	return &result, nil
}

// GetAlarm retrieves an alarm by ID.
func (c *Client) GetAlarm(ctx context.Context, id string) (*Alarm, error) {
	path := fmt.Sprintf("/alarm/alarms/%s", url.PathEscape(id))
	var result Alarm
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, fmt.Errorf("getting alarm %q: %w", id, err)
	}
	return &result, nil
}

// UpdateAlarm updates the mutable fields of an alarm (text, status, severity).
func (c *Client) UpdateAlarm(ctx context.Context, id, text, status, severity string) (*Alarm, error) {
	path := fmt.Sprintf("/alarm/alarms/%s", url.PathEscape(id))
	body := alarmUpdateRequest{
		Text:     text,
		Status:   status,
		Severity: severity,
	}
	var result Alarm
	if _, err := c.doJSON(ctx, http.MethodPut, path, body, &result); err != nil {
		return nil, fmt.Errorf("updating alarm %q: %w", id, err)
	}
	return &result, nil
}

// DeleteAlarm deletes an alarm by ID. Returns nil if not found.
func (c *Client) DeleteAlarm(ctx context.Context, id string) error {
	path := fmt.Sprintf("/alarm/alarms/%s", url.PathEscape(id))
	_, err := c.doJSON(ctx, http.MethodDelete, path, nil, nil)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("deleting alarm %q: %w", id, err)
	}
	return nil
}

// ListAlarms retrieves alarms with optional filters. All pages are followed.
func (c *Client) ListAlarms(ctx context.Context, sourceID, status, severity, alarmType string) ([]Alarm, error) {
	base := "/alarm/alarms?pageSize=100"
	if sourceID != "" {
		base += "&source=" + url.QueryEscape(sourceID)
	}
	if status != "" {
		base += "&status=" + url.QueryEscape(status)
	}
	if severity != "" {
		base += "&severity=" + url.QueryEscape(severity)
	}
	if alarmType != "" {
		base += "&type=" + url.QueryEscape(alarmType)
	}

	var all []Alarm
	path := base
	for path != "" {
		var page alarmCollectionPage
		if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &page); err != nil {
			return nil, fmt.Errorf("listing alarms: %w", err)
		}
		all = append(all, page.Alarms...)
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
