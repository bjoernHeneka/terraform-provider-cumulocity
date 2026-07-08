package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// Measurement represents a Cumulocity measurement.
type Measurement struct {
	ID           string            `json:"id,omitempty"`
	Type         string            `json:"type,omitempty"`
	Time         string            `json:"time,omitempty"`
	CreationTime string            `json:"creationTime,omitempty"`
	Self         string            `json:"self,omitempty"`
	Source       MeasurementSource `json:"source"`
	// FragmentsJSON holds the serialised measurement fragment data (everything
	// except the core fields above). It is not a JSON field — it is populated
	// by measurementFromRaw.
	FragmentsJSON string `json:"-"`
}

// MeasurementSource holds the managed object reference embedded in a measurement.
type MeasurementSource struct {
	ID string `json:"id"`
}

var knownMeasurementFields = map[string]bool{
	"id": true, "type": true, "time": true,
	"creationTime": true, "self": true, "source": true,
}

// measurementFromRaw converts a raw JSON map (as returned by the API) into a
// Measurement, placing all non-core keys into FragmentsJSON.
func measurementFromRaw(raw map[string]json.RawMessage) (*Measurement, error) {
	m := &Measurement{}
	if v, ok := raw["id"]; ok {
		_ = json.Unmarshal(v, &m.ID)
	}
	if v, ok := raw["type"]; ok {
		_ = json.Unmarshal(v, &m.Type)
	}
	if v, ok := raw["time"]; ok {
		_ = json.Unmarshal(v, &m.Time)
	}
	if v, ok := raw["creationTime"]; ok {
		_ = json.Unmarshal(v, &m.CreationTime)
	}
	if v, ok := raw["self"]; ok {
		_ = json.Unmarshal(v, &m.Self)
	}
	if v, ok := raw["source"]; ok {
		_ = json.Unmarshal(v, &m.Source)
	}

	fragments := map[string]json.RawMessage{}
	for k, v := range raw {
		if !knownMeasurementFields[k] {
			fragments[k] = v
		}
	}
	if len(fragments) > 0 {
		b, _ := json.Marshal(fragments)
		m.FragmentsJSON = string(b)
	} else {
		m.FragmentsJSON = "{}"
	}
	return m, nil
}

// CreateMeasurement posts a new measurement to /measurement/measurements.
// fragmentsJSON is a JSON object string containing the fragment data
// (e.g. {"c8y_Temperature":{"T":{"value":22.5,"unit":"°C"}}}); pass "" or "{}" for none.
func (c *Client) CreateMeasurement(ctx context.Context, sourceID, measType, measTime, fragmentsJSON string) (*Measurement, error) {
	body := map[string]interface{}{}
	if fragmentsJSON != "" && fragmentsJSON != "{}" {
		if err := json.Unmarshal([]byte(fragmentsJSON), &body); err != nil {
			return nil, fmt.Errorf("parsing measurement fragments: %w", err)
		}
	}
	body["source"] = map[string]string{"id": sourceID}
	body["type"] = measType
	body["time"] = measTime

	var raw map[string]json.RawMessage
	if _, err := c.doJSON(ctx, http.MethodPost, "/measurement/measurements", body, &raw); err != nil {
		return nil, fmt.Errorf("creating measurement: %w", err)
	}
	return measurementFromRaw(raw)
}

// GetMeasurement retrieves a single measurement by ID.
func (c *Client) GetMeasurement(ctx context.Context, id string) (*Measurement, error) {
	path := fmt.Sprintf("/measurement/measurements/%s", url.PathEscape(id))
	var raw map[string]json.RawMessage
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &raw); err != nil {
		return nil, fmt.Errorf("getting measurement %q: %w", id, err)
	}
	return measurementFromRaw(raw)
}

// DeleteMeasurement removes a measurement by ID.
func (c *Client) DeleteMeasurement(ctx context.Context, id string) error {
	path := fmt.Sprintf("/measurement/measurements/%s", url.PathEscape(id))
	_, err := c.doJSON(ctx, http.MethodDelete, path, nil, nil)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("deleting measurement %q: %w", id, err)
	}
	return nil
}

type measurementCollectionPage struct {
	Measurements []Measurement `json:"measurements"`
	Next         string        `json:"next,omitempty"`
}

// ListMeasurements retrieves measurements with optional filters. All pages are followed.
func (c *Client) ListMeasurements(ctx context.Context, sourceID, measType, dateFrom, dateTo, valueFragmentType, valueFragmentSeries string) ([]Measurement, error) {
	base := "/measurement/measurements?pageSize=100"
	if sourceID != "" {
		base += "&source=" + url.QueryEscape(sourceID)
	}
	if measType != "" {
		base += "&type=" + url.QueryEscape(measType)
	}
	if dateFrom != "" {
		base += "&dateFrom=" + url.QueryEscape(dateFrom)
	}
	if dateTo != "" {
		base += "&dateTo=" + url.QueryEscape(dateTo)
	}
	if valueFragmentType != "" {
		base += "&valueFragmentType=" + url.QueryEscape(valueFragmentType)
	}
	if valueFragmentSeries != "" {
		base += "&valueFragmentSeries=" + url.QueryEscape(valueFragmentSeries)
	}

	var all []Measurement
	path := base
	for path != "" {
		var page measurementCollectionPage
		if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &page); err != nil {
			return nil, fmt.Errorf("listing measurements: %w", err)
		}
		all = append(all, page.Measurements...)
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
