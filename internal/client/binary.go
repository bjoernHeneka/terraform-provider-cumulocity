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
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
)

// Binary represents a Cumulocity inventory binary (file managed object).
type Binary struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Type        string `json:"type,omitempty"`
	ContentType string `json:"contentType,omitempty"`
	Length      int64  `json:"length,omitempty"`
	Owner       string `json:"owner,omitempty"`
	Self        string `json:"self,omitempty"`
	LastUpdated string `json:"lastUpdated,omitempty"`
}

// binaryObjectInfo is the JSON `object` part of the multipart upload.
type binaryObjectInfo struct {
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
}

// UploadBinary uploads a file to /inventory/binaries via multipart form data.
// name defaults to the base filename if empty; contentType defaults to application/octet-stream.
func (c *Client) UploadBinary(ctx context.Context, filePath, name, contentType string) (*Binary, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening binary file %q: %w", filePath, err)
	}
	defer f.Close()

	if name == "" {
		name = filepath.Base(filePath)
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	objectJSON, err := json.Marshal(binaryObjectInfo{Name: name, Type: contentType})
	if err != nil {
		return nil, fmt.Errorf("marshalling binary object: %w", err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Write the `object` field as plain text (not a file).
	objectField, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": {`form-data; name="object"`},
		"Content-Type":        {"application/json"},
	})
	if err != nil {
		return nil, fmt.Errorf("creating object part: %w", err)
	}
	if _, err := objectField.Write(objectJSON); err != nil {
		return nil, fmt.Errorf("writing object part: %w", err)
	}

	// Write the `file` field with the binary content.
	filePart, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("creating file part: %w", err)
	}
	if _, err := io.Copy(filePart, f); err != nil {
		return nil, fmt.Errorf("copying file content: %w", err)
	}
	writer.Close()

	uploadURL := c.BaseURL + "/inventory/binaries"
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
		return nil, fmt.Errorf("uploading binary: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading upload response: %w", err)
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("upload failed (HTTP %d): %s", resp.StatusCode, truncateBody(respBody))
	}
	var result Binary
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parsing upload response: %w", err)
	}
	return &result, nil
}

// GetBinary retrieves binary metadata from the inventory managed object endpoint.
// The binary GET endpoint returns the raw file; use the managed object endpoint for metadata.
func (c *Client) GetBinary(ctx context.Context, id string) (*Binary, error) {
	path := fmt.Sprintf("/inventory/managedObjects/%s", url.PathEscape(id))
	var result Binary
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, fmt.Errorf("getting binary %q: %w", id, err)
	}
	return &result, nil
}

// ReplaceBinary replaces the file content of an existing binary.
// Only the content is replaced; name and contentType remain unchanged.
func (c *Client) ReplaceBinary(ctx context.Context, id, filePath string) (*Binary, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening binary file %q: %w", filePath, err)
	}
	defer f.Close()

	replaceURL := fmt.Sprintf("%s/inventory/binaries/%s", c.BaseURL, url.PathEscape(id))
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, replaceURL, f)
	if err != nil {
		return nil, fmt.Errorf("creating replace request: %w", err)
	}
	req.Header.Set("Authorization", "Basic "+c.encodedCreds)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Accept", "application/json")
	c.setUserAgent(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("replacing binary: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading replace response: %w", err)
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("replace failed (HTTP %d): %s", resp.StatusCode, truncateBody(respBody))
	}
	var result Binary
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parsing replace response: %w", err)
	}
	return &result, nil
}

// binaryCollectionPage is the JSON envelope returned by GET /inventory/binaries.
type binaryCollectionPage struct {
	ManagedObjects []Binary `json:"managedObjects"`
	Next           string   `json:"next,omitempty"`
}

// ListBinaries retrieves binary metadata with optional filters. All pages are followed.
// Pass empty strings to omit a filter.
func (c *Client) ListBinaries(ctx context.Context, owner, binaryType string) ([]Binary, error) {
	base := "/inventory/binaries?pageSize=100"
	if owner != "" {
		base += "&owner=" + url.QueryEscape(owner)
	}
	if binaryType != "" {
		base += "&type=" + url.QueryEscape(binaryType)
	}

	var all []Binary
	path := base
	for path != "" {
		var page binaryCollectionPage
		if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &page); err != nil {
			return nil, fmt.Errorf("listing binaries: %w", err)
		}
		all = append(all, page.ManagedObjects...)
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

// DeleteBinary removes a binary managed object and its stored file.
func (c *Client) DeleteBinary(ctx context.Context, id string) error {
	path := fmt.Sprintf("/inventory/binaries/%s", url.PathEscape(id))
	_, err := c.doJSON(ctx, http.MethodDelete, path, nil, nil)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("deleting binary %q: %w", id, err)
	}
	return nil
}
