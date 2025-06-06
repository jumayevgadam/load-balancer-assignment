package selector

import (
	"context"
	"net/url"
	"sync/atomic"

	"github.com/jumayevgadam/golb"
)

// BasicBalancer is a simple round-robin load balancer.
type BasicBalancer struct {
	backends []*backendImpl // list of backends.
	counter  uint32         // keeps track of next backend index.
}

// NewBasicLoadBalancer creates a new BasicBalancer instance.
// It takes a list of backend URLs and initializes backendImpls for each of them.
func NewBasicLoadBalancer(urls []*url.URL) *BasicBalancer {
	backends := make([]*backendImpl, len(urls))
	for i, u := range urls {
		backends[i] = &backendImpl{addr: u.Host}
	}

	return &BasicBalancer{backends: backends}
}

// NextServer selects the next backend server using round-robin strategy and uses atomic counter to ensure thread-safe selection under concurrency.
func (b *BasicBalancer) NextServer() *backendImpl {
	if len(b.backends) == 0 {
		return nil // no servers available
	}
	// Add 1 to the counter and choose a backend index in a repeating way.
	idx := int(atomic.AddUint32(&b.counter, 1)-1) % len(b.backends)

	return b.backends[idx]
}

// Invoke sends the request to the next available backend and returns an error if no backends are registered.
func (b *BasicBalancer) Invoke(ctx context.Context, req golb.Request) (golb.Response, error) {
	backend := b.NextServer()
	if backend == nil {
		return nil, ErrBackendServersEmpty
	}

	return backend.Invoke(ctx, req)
}
