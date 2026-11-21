package push

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Client struct {
	endpoint   string
	httpClient *http.Client
}

func (c *Client) CloseIdleConnections() {
	c.httpClient.CloseIdleConnections()
}

func New(baseURL string, timeout time.Duration) *Client {
	return &Client{
		endpoint: strings.TrimRight(baseURL, "/") + "/api/v1/push",
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				DialContext:         (&net.Dialer{Timeout: timeout}).DialContext,
				TLSHandshakeTimeout: timeout,
				IdleConnTimeout:     90 * time.Second,
			},
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
	}
}

func (c *Client) RequestPosition(ctx context.Context, driverID, tripID uuid.UUID) error {
	raw := map[string]any{
		"request_id":   uuid.New().String(),
		"recipient_id": driverID.String(),
		"kind":         "REQUEST_POSITION",
		"data": map[string]any{
			"trip_id": tripID.String(),
		},
	}
	body, err := json.Marshal(raw)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send position request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("push service returned HTTP %d", resp.StatusCode)
	}
	return nil
}
