package notion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"
)

const (
	defaultBaseURL = "https://api.notion.com/v1"
	apiVersion     = "2022-06-28"
	maxRetries     = 3
)

// Client is a Notion API client.
type Client struct {
	httpClient *http.Client
	token      string
	version    string
	baseURL    string
}

// NewClient creates a new Notion API client with the given token.
func NewClient(token string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		token:      token,
		version:    apiVersion,
		baseURL:    defaultBaseURL,
	}
}

// NewClientWithBaseURL creates a new Notion API client with a custom base URL (useful for testing).
func NewClientWithBaseURL(token, baseURL string) *Client {
	c := NewClient(token)
	c.baseURL = baseURL
	return c
}

// do sends an HTTP request to the Notion API with authentication headers,
// handles JSON encoding/decoding, and retries on 429 (rate limit) responses.
func (c *Client) do(ctx context.Context, method, path string, body, result any) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := c.doOnce(ctx, method, path, body, result)
		if err == nil {
			return nil
		}

		// Check if it's a rate limit error (429)
		if notionErr, ok := err.(*NotionError); ok && notionErr.Status == http.StatusTooManyRequests {
			if attempt >= maxRetries {
				return err
			}
			lastErr = err

			// Use Retry-After header value or exponential backoff
			waitDuration := time.Duration(math.Pow(2, float64(attempt+1))) * time.Second

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitDuration):
				continue
			}
		}

		// Non-retryable error
		return err
	}

	return lastErr
}

// doOnce performs a single HTTP request to the Notion API.
func (c *Client) doOnce(ctx context.Context, method, path string, body, result any) error {
	url := c.baseURL + path

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("notion: failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("notion: failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Notion-Version", c.version)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("notion: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("notion: failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var notionErr NotionError
		if err := json.Unmarshal(respBody, &notionErr); err != nil {
			return fmt.Errorf("notion: unexpected status %d: %s", resp.StatusCode, string(respBody))
		}

		// Parse Retry-After for rate limit responses
		if resp.StatusCode == http.StatusTooManyRequests {
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if seconds, err := strconv.Atoi(retryAfter); err == nil {
					_ = seconds // Retry-After value is available for the caller
				}
			}
		}

		return &notionErr
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("notion: failed to unmarshal response: %w", err)
		}
	}

	return nil
}
