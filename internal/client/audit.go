package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// AuditRecord represents a Cumulocity audit record.
type AuditRecord struct {
	ID           string      `json:"id,omitempty"`
	Activity     string      `json:"activity,omitempty"`
	Application  string      `json:"application,omitempty"`
	CreationTime string      `json:"creationTime,omitempty"`
	Severity     string      `json:"severity,omitempty"`
	Source       AuditSource `json:"source"`
	Text         string      `json:"text,omitempty"`
	Time         string      `json:"time,omitempty"`
	Type         string      `json:"type,omitempty"`
	User         string      `json:"user,omitempty"`
	Self         string      `json:"self,omitempty"`
}

// AuditSource is the platform component reference in an audit record.
type AuditSource struct {
	ID string `json:"id"`
}

type auditRecordCreateRequest struct {
	Activity string      `json:"activity"`
	Source   AuditSource `json:"source"`
	Text     string      `json:"text"`
	Time     string      `json:"time"`
	Type     string      `json:"type"`
	User     string      `json:"user,omitempty"`
}

type auditRecordCollectionPage struct {
	AuditRecords []AuditRecord `json:"auditRecords"`
	Next         string        `json:"next,omitempty"`
}

// CreateAuditRecord creates a new audit record.
func (c *Client) CreateAuditRecord(ctx context.Context, sourceID, activity, text, time, recordType, user string) (*AuditRecord, error) {
	body := auditRecordCreateRequest{
		Activity: activity,
		Source:   AuditSource{ID: sourceID},
		Text:     text,
		Time:     time,
		Type:     recordType,
		User:     user,
	}
	var result AuditRecord
	if _, err := c.doJSON(ctx, http.MethodPost, "/audit/auditRecords", body, &result); err != nil {
		return nil, fmt.Errorf("creating audit record: %w", err)
	}
	return &result, nil
}

// GetAuditRecord retrieves an audit record by ID.
func (c *Client) GetAuditRecord(ctx context.Context, id string) (*AuditRecord, error) {
	path := fmt.Sprintf("/audit/auditRecords/%s", url.PathEscape(id))
	var result AuditRecord
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, fmt.Errorf("getting audit record %q: %w", id, err)
	}
	return &result, nil
}

// ListAuditRecords retrieves audit records with optional filters. All pages are followed.
func (c *Client) ListAuditRecords(ctx context.Context, sourceID, recordType, user, application string) ([]AuditRecord, error) {
	base := "/audit/auditRecords?pageSize=100"
	if sourceID != "" {
		base += "&source=" + url.QueryEscape(sourceID)
	}
	if recordType != "" {
		base += "&type=" + url.QueryEscape(recordType)
	}
	if user != "" {
		base += "&user=" + url.QueryEscape(user)
	}
	if application != "" {
		base += "&application=" + url.QueryEscape(application)
	}

	var all []AuditRecord
	path := base
	for path != "" {
		var page auditRecordCollectionPage
		if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &page); err != nil {
			return nil, fmt.Errorf("listing audit records: %w", err)
		}
		all = append(all, page.AuditRecords...)
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
