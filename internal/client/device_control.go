package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// ─── Operation ────────────────────────────────────────────────────────────────

// Operation represents a Cumulocity device control operation.
type Operation struct {
	ID              string `json:"id,omitempty"`
	DeviceID        string `json:"deviceId,omitempty"`
	Status          string `json:"status,omitempty"`
	FailureReason   string `json:"failureReason,omitempty"`
	CreationTime    string `json:"creationTime,omitempty"`
	BulkOperationID int64  `json:"bulkOperationId,omitempty"`
	Self            string `json:"self,omitempty"`
}

type operationCollectionPage struct {
	Operations []Operation `json:"operations"`
	Next       string      `json:"next,omitempty"`
}

// CreateOperation posts a new operation. fragmentsJSON is a JSON object with
// the operation payload (e.g. {"c8y_Restart":{}}); it is merged with deviceId
// and the optional description.
func (c *Client) CreateOperation(ctx context.Context, deviceID, description, fragmentsJSON string) (*Operation, error) {
	body := make(map[string]json.RawMessage)
	if fragmentsJSON != "" {
		if err := json.Unmarshal([]byte(fragmentsJSON), &body); err != nil {
			return nil, fmt.Errorf("parsing fragments_json: %w", err)
		}
	}
	deviceIDBytes, _ := json.Marshal(deviceID)
	body["deviceId"] = deviceIDBytes
	if description != "" {
		descBytes, _ := json.Marshal(description)
		body["description"] = descBytes
	}

	var result Operation
	if _, err := c.doJSON(ctx, http.MethodPost, "/devicecontrol/operations", body, &result); err != nil {
		return nil, fmt.Errorf("creating operation on device %q: %w", deviceID, err)
	}
	return &result, nil
}

// GetOperation retrieves a device operation by ID.
func (c *Client) GetOperation(ctx context.Context, id string) (*Operation, error) {
	path := fmt.Sprintf("/devicecontrol/operations/%s", url.PathEscape(id))
	var result Operation
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, fmt.Errorf("getting operation %q: %w", id, err)
	}
	return &result, nil
}

// CancelOperation attempts to cancel a PENDING operation by setting its status
// to FAILED. This is the closest Cumulocity offers to an individual delete.
func (c *Client) CancelOperation(ctx context.Context, id string) error {
	op, err := c.GetOperation(ctx, id)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if op.Status != "PENDING" {
		// Already executing or done — nothing to cancel.
		return nil
	}
	path := fmt.Sprintf("/devicecontrol/operations/%s", url.PathEscape(id))
	body := map[string]string{"status": "FAILED", "failureReason": "cancelled by Terraform"}
	if _, err := c.doJSON(ctx, http.MethodPut, path, body, nil); err != nil {
		return fmt.Errorf("cancelling operation %q: %w", id, err)
	}
	return nil
}

// ListOperations retrieves operations filtered by optional deviceID and status.
func (c *Client) ListOperations(ctx context.Context, deviceID, status string) ([]Operation, error) {
	base := "/devicecontrol/operations?pageSize=100"
	if deviceID != "" {
		base += "&deviceId=" + url.QueryEscape(deviceID)
	}
	if status != "" {
		base += "&status=" + url.QueryEscape(status)
	}

	var all []Operation
	path := base
	for path != "" {
		var page operationCollectionPage
		if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &page); err != nil {
			return nil, fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, page.Operations...)
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

// ─── New Device Request ───────────────────────────────────────────────────────

// NewDeviceRequest represents a Cumulocity device registration request.
type NewDeviceRequest struct {
	ID           string `json:"id,omitempty"`
	GroupID      string `json:"groupId,omitempty"`
	Type         string `json:"type,omitempty"`
	TenantID     string `json:"tenantId,omitempty"`
	Status       string `json:"status,omitempty"`
	Owner        string `json:"owner,omitempty"`
	CreationTime string `json:"creationTime,omitempty"`
	Self         string `json:"self,omitempty"`
}

type newDeviceRequestCreate struct {
	ID            string `json:"id"`
	GroupID       string `json:"groupId,omitempty"`
	Type          string `json:"type,omitempty"`
	SecurityToken string `json:"securityToken,omitempty"`
}

type newDeviceRequestUpdate struct {
	Status        string `json:"status"`
	SecurityToken string `json:"securityToken,omitempty"`
}

// CreateNewDeviceRequest registers a new device. If a request for the device
// already exists (HTTP 422), the existing request is returned instead.
func (c *Client) CreateNewDeviceRequest(ctx context.Context, deviceID, groupID, deviceType, securityToken string) (*NewDeviceRequest, error) {
	body := newDeviceRequestCreate{
		ID:            deviceID,
		GroupID:       groupID,
		Type:          deviceType,
		SecurityToken: securityToken,
	}
	var result NewDeviceRequest
	_, err := c.doJSON(ctx, http.MethodPost, "/devicecontrol/newDeviceRequests", body, &result)
	if err != nil {
		// 422 means the request already exists — adopt it.
		if strings.Contains(err.Error(), "422") {
			return c.GetNewDeviceRequest(ctx, deviceID)
		}
		return nil, fmt.Errorf("creating new device request %q: %w", deviceID, err)
	}
	return &result, nil
}

// GetNewDeviceRequest retrieves a device registration request by device ID.
func (c *Client) GetNewDeviceRequest(ctx context.Context, deviceID string) (*NewDeviceRequest, error) {
	path := fmt.Sprintf("/devicecontrol/newDeviceRequests/%s", url.PathEscape(deviceID))
	var result NewDeviceRequest
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, fmt.Errorf("getting new device request %q: %w", deviceID, err)
	}
	return &result, nil
}

// UpdateNewDeviceRequest updates the status of a device registration request.
func (c *Client) UpdateNewDeviceRequest(ctx context.Context, deviceID, status, securityToken string) (*NewDeviceRequest, error) {
	path := fmt.Sprintf("/devicecontrol/newDeviceRequests/%s", url.PathEscape(deviceID))
	body := newDeviceRequestUpdate{Status: status, SecurityToken: securityToken}
	var result NewDeviceRequest
	if _, err := c.doJSON(ctx, http.MethodPut, path, body, &result); err != nil {
		return nil, fmt.Errorf("updating new device request %q: %w", deviceID, err)
	}
	return &result, nil
}

// DeleteNewDeviceRequest removes a device registration request.
func (c *Client) DeleteNewDeviceRequest(ctx context.Context, deviceID string) error {
	path := fmt.Sprintf("/devicecontrol/newDeviceRequests/%s", url.PathEscape(deviceID))
	_, err := c.doJSON(ctx, http.MethodDelete, path, nil, nil)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("deleting new device request %q: %w", deviceID, err)
	}
	return nil
}

// ─── Device Credentials ───────────────────────────────────────────────────────

// DeviceCredentials holds the auto-generated credentials for a device.
type DeviceCredentials struct {
	ID            string `json:"id,omitempty"`
	Username      string `json:"username,omitempty"`
	Password      string `json:"password,omitempty"`
	TenantID      string `json:"tenantId,omitempty"`
	Self          string `json:"self,omitempty"`
	SecurityToken string `json:"securityToken,omitempty"`
}

// CreateDeviceCredentials requests auto-generated credentials for a device.
func (c *Client) CreateDeviceCredentials(ctx context.Context, deviceID, securityToken string) (*DeviceCredentials, error) {
	body := map[string]string{"id": deviceID}
	if securityToken != "" {
		body["securityToken"] = securityToken
	}
	var result DeviceCredentials
	if _, err := c.doJSON(ctx, http.MethodPost, "/devicecontrol/deviceCredentials", body, &result); err != nil {
		return nil, fmt.Errorf("creating device credentials for %q: %w", deviceID, err)
	}
	return &result, nil
}

// ─── Bulk Operation ───────────────────────────────────────────────────────────

// BulkOperation represents a Cumulocity bulk operation targeting a device group.
type BulkOperation struct {
	ID                 string                 `json:"id,omitempty"`
	GroupID            string                 `json:"groupId,omitempty"`
	FailedParentID     string                 `json:"failedParentId,omitempty"`
	StartDate          string                 `json:"startDate,omitempty"`
	CreationRamp       float64                `json:"creationRamp,omitempty"`
	OperationPrototype json.RawMessage        `json:"operationPrototype,omitempty"`
	Status             string                 `json:"status,omitempty"`
	GeneralStatus      string                 `json:"generalStatus,omitempty"`
	Progress           *BulkOperationProgress `json:"progress,omitempty"`
	Self               string                 `json:"self,omitempty"`
}

// BulkOperationProgress holds counters for a bulk operation.
type BulkOperationProgress struct {
	Pending    int `json:"pending"`
	Failed     int `json:"failed"`
	Executing  int `json:"executing"`
	Successful int `json:"successful"`
	All        int `json:"all"`
}

type bulkOperationCreateRequest struct {
	GroupID            string          `json:"groupId,omitempty"`
	FailedParentID     string          `json:"failedParentId,omitempty"`
	StartDate          string          `json:"startDate"`
	CreationRamp       float64         `json:"creationRamp"`
	OperationPrototype json.RawMessage `json:"operationPrototype"`
}

type bulkOperationUpdateRequest struct {
	StartDate          string          `json:"startDate,omitempty"`
	CreationRamp       float64         `json:"creationRamp,omitempty"`
	OperationPrototype json.RawMessage `json:"operationPrototype,omitempty"`
}

// CreateBulkOperation creates a new bulk operation.
func (c *Client) CreateBulkOperation(ctx context.Context, groupID, failedParentID, startDate string, creationRamp float64, operationPrototypeJSON string) (*BulkOperation, error) {
	proto := json.RawMessage("{}")
	if operationPrototypeJSON != "" {
		proto = json.RawMessage(operationPrototypeJSON)
	}
	body := bulkOperationCreateRequest{
		GroupID:            groupID,
		FailedParentID:     failedParentID,
		StartDate:          startDate,
		CreationRamp:       creationRamp,
		OperationPrototype: proto,
	}
	var result BulkOperation
	if _, err := c.doJSON(ctx, http.MethodPost, "/devicecontrol/bulkoperations", body, &result); err != nil {
		return nil, fmt.Errorf("creating bulk operation: %w", err)
	}
	return &result, nil
}

// GetBulkOperation retrieves a bulk operation by ID.
func (c *Client) GetBulkOperation(ctx context.Context, id string) (*BulkOperation, error) {
	path := fmt.Sprintf("/devicecontrol/bulkoperations/%s", url.PathEscape(id))
	var result BulkOperation
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, fmt.Errorf("getting bulk operation %q: %w", id, err)
	}
	return &result, nil
}

// UpdateBulkOperation updates scheduling and prototype of a bulk operation.
func (c *Client) UpdateBulkOperation(ctx context.Context, id, startDate string, creationRamp float64, operationPrototypeJSON string) (*BulkOperation, error) {
	path := fmt.Sprintf("/devicecontrol/bulkoperations/%s", url.PathEscape(id))
	proto := json.RawMessage("{}")
	if operationPrototypeJSON != "" {
		proto = json.RawMessage(operationPrototypeJSON)
	}
	body := bulkOperationUpdateRequest{
		StartDate:          startDate,
		CreationRamp:       creationRamp,
		OperationPrototype: proto,
	}
	var result BulkOperation
	if _, err := c.doJSON(ctx, http.MethodPut, path, body, &result); err != nil {
		return nil, fmt.Errorf("updating bulk operation %q: %w", id, err)
	}
	return &result, nil
}

// DeleteBulkOperation removes a bulk operation. Returns nil if already gone.
func (c *Client) DeleteBulkOperation(ctx context.Context, id string) error {
	path := fmt.Sprintf("/devicecontrol/bulkoperations/%s", url.PathEscape(id))
	_, err := c.doJSON(ctx, http.MethodDelete, path, nil, nil)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("deleting bulk operation %q: %w", id, err)
	}
	return nil
}
