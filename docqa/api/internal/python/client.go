package python

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	base string
	http *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		base: baseURL,
		http: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *Client) Ingest(ctx context.Context, docID string, chunks []string) error {
	body, _ := json.Marshal(map[string]any{
		"doc_id": docID,
		"chunks": chunks,
	})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/ingest", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("ingest request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ingest returned %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) Retrieve(ctx context.Context, docID, query string) ([]string, error) {
	body, _ := json.Marshal(map[string]any{
		"doc_id": docID,
		"query":  query,
	})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/retrieve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("retrieve request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("retrieve returned %d", resp.StatusCode)
	}

	var result struct {
		Chunks []string `json:"chunks"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode retrieve response: %w", err)
	}
	return result.Chunks, nil
}
