package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"agro-subsidy/go-service/internal/model"
)

// MLClient calls the FastAPI /score endpoint with exponential-ish retry.
type MLClient struct {
	baseURL    string
	httpClient *http.Client
	maxRetries int
}

func NewMLClient(baseURL string) *MLClient {
	return &MLClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxRetries: 3,
	}
}

func (c *MLClient) Score(ctx context.Context, req model.MLRequest) (*model.MLResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < c.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt) * 500 * time.Millisecond
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		resp, err := c.doRequest(ctx, body)
		if err == nil {
			return resp, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("all %d retries exhausted: %w", c.maxRetries, lastErr)
}

func (c *MLClient) doRequest(ctx context.Context, body []byte) (*model.MLResponse, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, c.baseURL+"/score", bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ml-service HTTP %d", resp.StatusCode)
	}

	var result model.MLResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &result, nil
}
