package legacyAPI

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"checkin.service/internal/ports/messaging"
)

// LegacyAPIClient contract for check-out system
type LegacyAPIClient interface {
	RecordCheckOut(ctx context.Context, event messaging.CheckOutEvent) error
}

// HTTPClient API client using HTTP
type HTTPClient struct {
	client  *http.Client
	baseURL string
}

// NewHTTPClient new HTTPClient
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
	}
}

// RecordCheckOut send the check-out event to legacy API
func (c *HTTPClient) RecordCheckOut(ctx context.Context, event messaging.CheckOutEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal legacy api payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create legacy api request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call legacy api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("legacy api returned non-successful status code: %d", resp.StatusCode)
	}

	log.Printf("Successfully recorded check-out for EmployeeID %s in legacy system", event.EmployeeID)
	return nil
}
