package golb

import (
	"context"
	"net/url"
	"sync/atomic"
)

// Ensure IntermediateBalancer implements the Backend interface.
var _ Backend = (*IntermediateBalancer)(nil)

// IntermediateBalancer adds failure tracking and re-inclusion.
type IntermediateBalancer struct {
	backends []*BackendImpl // List of all backends.
	counter  uint32         // Atomic counter for round-robin index tracking.
}

// NewIntermediateLoadBalancer creates and returns a new initialied IntermediateBalancer.
func NewIntermediateLoadBalancer(urls []*url.URL) *IntermediateBalancer {
	backends := make([]*BackendImpl, len(urls))
	for i, u := range urls {
		backends[i] = NewBackend(u.Host)
	}

	return &IntermediateBalancer{backends: backends}
}

// GetNextServer selects the next healthy backend using round-robin strategy.
// Skips unhealthy backends, tries at most once per backend.
func (b *IntermediateBalancer) GetNextServer() *BackendImpl {
	n := len(b.backends)
	if n == 0 {
		return nil
	}
	// Try up to n times to find a healthy backend.
	for i := 0; i < n; i++ {
		// Atomically get the next index in round-robin fashion.
		idx := int(atomic.AddUint32(&b.counter, 1)-1) % n
		backend := b.backends[idx]

		if backend.IsHealthy() {
			return backend
		}
	}

	return nil // No healthy backends found.
}

// Invoke sends the request to a healthy backend.
// If the backend fails repeatedly, it's marked unhealthy.
func (b *IntermediateBalancer) Invoke(ctx context.Context, req Request) (Response, error) {
	backend := b.GetNextServer()
	if backend == nil {
		return nil, ErrBackendServersEmpty
	}

	resp, err := backend.Invoke(ctx, req)
	if err != nil {
		// Check failure count and mark unhealthy if maxFailure reached.
		backend.mu.Lock()
		if backend.failureCount >= maxFailureCount { // Max 3 consecutive failures.
			backend.MarkUnhealthy()
		}
		backend.mu.Unlock()

		return nil, err
	}

	return resp, nil
}
