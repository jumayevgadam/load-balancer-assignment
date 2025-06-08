package golb

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const (
	httpClientTimeOut = 5 * time.Second

	maxFailureCount = 3
)

type (
	Request  interface{}
	Response interface{}
)

type Backend interface {
	Invoke(ctx context.Context, req Request) (Response, error)
}

type BackendImpl struct {
	addr           string       // addr general need for all level.
	failureCount   int32        // track consecutive failures (for intermediate level).
	lastFailure    time.Time    // time of last failure (for intermediate level).
	healthy        int32        // 1 if healthy, 0 if unhealthy (for intermediate level).
	activeRequests int64        // for tracking current active requests.
	mu             sync.Mutex   // protects failureCount and lastFailure.
	client         *http.Client // HTTP client for real requests.
}

// Ensure BackendImpl implements Backend.
var _ Backend = (*BackendImpl)(nil)

// NewBackend creates a new BackendImpl.
func NewBackend(addr string) *BackendImpl {
	return &BackendImpl{
		addr:    addr,
		healthy: 1, // start as healthy.
		client: &http.Client{
			Timeout: httpClientTimeOut,
		},
	}
}

func (b *BackendImpl) Invoke(ctx context.Context, req Request) (Response, error) {
	// for advanced level we need to track active requests.
	atomic.AddInt64(&b.activeRequests, 1)
	defer atomic.AddInt64(&b.activeRequests, -1)

	// create http request.
	url := fmt.Sprintf("https://%s", b.addr)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	// send request.
	resp, err := b.client.Do(httpReq)
	b.mu.Lock()
	defer b.mu.Unlock()
	// handle failures
	if err != nil || resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		b.failureCount++
		b.lastFailure = time.Now()

		if resp != nil {
			resp.Body.Close()
		}

		if err != nil {
			return nil, fmt.Errorf("backend %s failed: %w", b.addr, err)
		}

		return nil, fmt.Errorf("backend %s returned status %d", b.addr, resp.StatusCode)
	}
	defer resp.Body.Close()

	return fmt.Sprintf("Success from %s", b.addr), nil
}

// Host returns the backend's address.
func (b *BackendImpl) Host() string {
	return b.addr
}

// IsHealthy checks if the backend is healthy.
func (b *BackendImpl) IsHealthy() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if atomic.LoadInt32(&b.healthy) == 1 {
		return true
	}

	// recover backend after some recovery time.
	if time.Since(b.lastFailure) > 2*time.Second {
		log.Printf("backend %s recovered", b.addr)
		atomic.StoreInt32(&b.healthy, 1)
		b.failureCount = 0

		return true
	}

	return false
}

// MarkUnhealthy marks the backend as unhealthy.
func (b *BackendImpl) MarkUnhealthy() {
	atomic.StoreInt32(&b.healthy, 0)
}

// GetLoad returns the current number of active requests.
func (b *BackendImpl) GetLoad() int64 {
	return atomic.LoadInt64(&b.activeRequests)
}
