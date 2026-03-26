package fetch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetch_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-agent", r.Header.Get("User-Agent"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html><body><h1>Job Posting</h1><p>We are hiring.</p></body></html>"))
	}))
	defer server.Close()

	f := NewFetcher(5*time.Second, "test-agent")
	content, err := f.Fetch(context.Background(), server.URL)
	require.NoError(t, err)
	assert.Contains(t, content, "Job Posting")
	assert.Contains(t, content, "We are hiring")
}

func TestFetch_404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	f := NewFetcher(5*time.Second, "")
	_, err := f.Fetch(context.Background(), server.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestFetch_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	f := NewFetcher(100*time.Millisecond, "")
	_, err := f.Fetch(context.Background(), server.URL)
	require.Error(t, err)
}

func TestFetch_DefaultUserAgent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, DefaultUserAgent, r.Header.Get("User-Agent"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	f := NewFetcher(5*time.Second, "")
	_, err := f.Fetch(context.Background(), server.URL)
	require.NoError(t, err)
}

func TestFetch_FollowsRedirect(t *testing.T) {
	final := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("final destination"))
	}))
	defer final.Close()

	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, final.URL, http.StatusMovedPermanently)
	}))
	defer redirect.Close()

	f := NewFetcher(5*time.Second, "")
	content, err := f.Fetch(context.Background(), redirect.URL)
	require.NoError(t, err)
	assert.Equal(t, "final destination", content)
}
