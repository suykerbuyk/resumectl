package fetch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefaultUserAgent is used when no custom user agent is configured.
const DefaultUserAgent = "Mozilla/5.0 (compatible; resumectl/1.0)"

// Fetcher retrieves content from URLs.
type Fetcher struct {
	Client    *http.Client
	UserAgent string
}

// NewFetcher creates a Fetcher with the given timeout and user agent.
func NewFetcher(timeout time.Duration, userAgent string) *Fetcher {
	if userAgent == "" {
		userAgent = DefaultUserAgent
	}
	return &Fetcher{
		Client: &http.Client{
			Timeout: timeout,
		},
		UserAgent: userAgent,
	}
}

// Fetch retrieves the content at the given URL and returns it as a string.
func (f *Fetcher) Fetch(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("fetch: create request: %w", err)
	}
	req.Header.Set("User-Agent", f.UserAgent)

	resp, err := f.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch: request %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch: %s returned status %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("fetch: read response: %w", err)
	}

	return string(body), nil
}
