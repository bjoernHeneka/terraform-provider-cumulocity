package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// Application represents a Cumulocity application.
type Application struct {
	ID              string            `json:"id,omitempty"`
	Key             string            `json:"key,omitempty"`
	Name            string            `json:"name,omitempty"`
	Type            string            `json:"type,omitempty"`
	ContextPath     string            `json:"contextPath,omitempty"`
	Availability    string            `json:"availability,omitempty"`
	Description     string            `json:"description,omitempty"`
	Self            string            `json:"self,omitempty"`
	ActiveVersionID string            `json:"activeVersionId,omitempty"`
	Owner           *ApplicationOwner `json:"owner,omitempty"`
}

// ApplicationOwner holds tenant ownership info embedded in an Application.
type ApplicationOwner struct {
	Tenant struct {
		ID string `json:"id,omitempty"`
	} `json:"tenant,omitempty"`
}

// applicationCreateRequest is the body for POST /application/applications.
type applicationCreateRequest struct {
	Key          string `json:"key"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	ContextPath  string `json:"contextPath,omitempty"`
	Availability string `json:"availability,omitempty"`
	Description  string `json:"description,omitempty"`
}

// applicationUpdateRequest is the body for PUT /application/applications/{id}.
// Type is intentionally omitted — it is readOnly on PUT.
type applicationUpdateRequest struct {
	Key          string `json:"key,omitempty"`
	Name         string `json:"name,omitempty"`
	ContextPath  string `json:"contextPath,omitempty"`
	Availability string `json:"availability,omitempty"`
	Description  string `json:"description,omitempty"`
}

// ApplicationBinary is metadata for a single uploaded application binary.
type ApplicationBinary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Length      int64  `json:"length"`
	Created     string `json:"created"`
	DownloadURL string `json:"downloadUrl"`
	ContextPath string `json:"contextPath"`
}

// applicationBinaries is the response body for GET /application/applications/{id}/binaries.
type applicationBinaries struct {
	Attachments []ApplicationBinary `json:"attachments"`
}

// applicationCollectionPage is used for paginated listing.
type applicationCollectionPage struct {
	Applications []Application `json:"applications"`
	Next         string        `json:"next,omitempty"`
}

// CreateApplication creates a new application.
func (c *Client) CreateApplication(ctx context.Context, key, name, appType, contextPath, availability, description string) (*Application, error) {
	body := applicationCreateRequest{
		Key:          key,
		Name:         name,
		Type:         appType,
		ContextPath:  contextPath,
		Availability: availability,
		Description:  description,
	}
	var result Application
	if _, err := c.doJSON(ctx, http.MethodPost, "/application/applications", body, &result); err != nil {
		return nil, fmt.Errorf("creating application %q: %w", name, err)
	}
	return &result, nil
}

// GetApplication retrieves an application by its ID.
func (c *Client) GetApplication(ctx context.Context, id string) (*Application, error) {
	path := fmt.Sprintf("/application/applications/%s", url.PathEscape(id))
	var result Application
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, fmt.Errorf("getting application %q: %w", id, err)
	}
	return &result, nil
}

// UpdateApplication updates a mutable application. Type cannot be changed.
func (c *Client) UpdateApplication(ctx context.Context, id, key, name, contextPath, availability, description string) (*Application, error) {
	path := fmt.Sprintf("/application/applications/%s", url.PathEscape(id))
	body := applicationUpdateRequest{
		Key:          key,
		Name:         name,
		ContextPath:  contextPath,
		Availability: availability,
		Description:  description,
	}
	var result Application
	if _, err := c.doJSON(ctx, http.MethodPut, path, body, &result); err != nil {
		return nil, fmt.Errorf("updating application %q: %w", id, err)
	}
	return &result, nil
}

// DeleteApplication removes an application. Returns nil if already gone.
func (c *Client) DeleteApplication(ctx context.Context, id string) error {
	path := fmt.Sprintf("/application/applications/%s?force=true", url.PathEscape(id))
	_, err := c.doJSON(ctx, http.MethodDelete, path, nil, nil)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("deleting application %q: %w", id, err)
	}
	return nil
}

// UploadApplicationBinary uploads a ZIP file to an existing application.
// Returns the updated Application (which includes the new activeVersionId).
func (c *Client) UploadApplicationBinary(ctx context.Context, appID, filePath string) (*Application, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening binary file %q: %w", filePath, err)
	}
	defer f.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("creating multipart field: %w", err)
	}
	if _, err := io.Copy(part, f); err != nil {
		return nil, fmt.Errorf("copying file content: %w", err)
	}
	writer.Close()

	uploadURL := fmt.Sprintf("%s/application/applications/%s/binaries", c.BaseURL, url.PathEscape(appID))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, &body)
	if err != nil {
		return nil, fmt.Errorf("creating upload request: %w", err)
	}
	req.Header.Set("Authorization", "Basic "+c.encodedCreds)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	c.setUserAgent(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("uploading application binary: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading upload response: %w", err)
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("upload failed (HTTP %d): %s", resp.StatusCode, truncateBody(respBody))
	}

	// Parse the returned application to get activeVersionId.
	var result Application
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("parsing upload response: %w", err)
		}
	}
	return &result, nil
}

// GetApplicationBinaries lists all binaries attached to an application.
func (c *Client) GetApplicationBinaries(ctx context.Context, appID string) ([]ApplicationBinary, error) {
	path := fmt.Sprintf("/application/applications/%s/binaries", url.PathEscape(appID))
	var result applicationBinaries
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, fmt.Errorf("listing binaries for application %q: %w", appID, err)
	}
	return result.Attachments, nil
}

// GetApplicationBinaryByID returns a single binary attachment by its ID, or ErrNotFound.
func (c *Client) GetApplicationBinaryByID(ctx context.Context, appID, binaryID string) (*ApplicationBinary, error) {
	binaries, err := c.GetApplicationBinaries(ctx, appID)
	if err != nil {
		return nil, err
	}
	for i := range binaries {
		if binaries[i].ID == binaryID {
			return &binaries[i], nil
		}
	}
	return nil, ErrNotFound
}

// DeleteApplicationBinary removes a specific binary from an application.
func (c *Client) DeleteApplicationBinary(ctx context.Context, appID, binaryID string) error {
	path := fmt.Sprintf("/application/applications/%s/binaries/%s", url.PathEscape(appID), url.PathEscape(binaryID))
	_, err := c.doJSON(ctx, http.MethodDelete, path, nil, nil)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("deleting binary %q from application %q: %w", binaryID, appID, err)
	}
	return nil
}

// ListApplications retrieves all applications, following pagination.
func (c *Client) ListApplications(ctx context.Context) ([]Application, error) {
	path := "/application/applications?pageSize=100"
	var all []Application
	for path != "" {
		var page applicationCollectionPage
		if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &page); err != nil {
			return nil, fmt.Errorf("listing applications: %w", err)
		}
		all = append(all, page.Applications...)
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

// GetApplicationsByName retrieves applications by name.
// Returns all applications matching the given name (which may be zero, one, or multiple).
func (c *Client) GetApplicationsByName(ctx context.Context, name string) ([]Application, error) {
	path := fmt.Sprintf("/application/applicationsByName/%s", url.PathEscape(name))
	var page applicationCollectionPage
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &page); err != nil {
		return nil, fmt.Errorf("getting applications by name %q: %w", name, err)
	}
	return page.Applications, nil
}
