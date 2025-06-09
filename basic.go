package golb

import (
	"context"
	"net/url"
	"sync/atomic"
)

// Ensure BasicBalancer implements the Backend interface.
var _ Backend = (*BasicBalancer)(nil)

// BasicBalancer uses round-robin to distribute requests across backends.
type BasicBalancer struct {
	backends []*BackendImpl // List of backend servers.
	counter  uint32         // Atomic counter for round-robin selection.
}

// NewBasicLoadBalancer creates a BasicBalancer from a list of backend URLs.
func NewBasicLoadBalancer(urls []*url.URL) *BasicBalancer {
	backends := make([]*BackendImpl, len(urls))
	for i, u := range urls {
		backends[i] = NewBackend(u.Host)
	}

	return &BasicBalancer{backends: backends}
}

// GetNextServer selects the next backend using round-robin.
// Does not check for health; always returns the next backend.
func (b *BasicBalancer) GetNextServer() *BackendImpl {
	if len(b.backends) == 0 {
		return nil // no servers available.
	}
	// getting next server using modulo len(b.backends).
	idx := int(atomic.AddUint32(&b.counter, 1)-1) % len(b.backends)

	return b.backends[idx]
}

// Invoke sends the request to the next backend.
func (b *BasicBalancer) Invoke(ctx context.Context, req Request) (Response, error) {
	backend := b.GetNextServer()
	if backend == nil {
		return nil, ErrBackendServersEmpty
	}

	return backend.Invoke(ctx, req)
}
