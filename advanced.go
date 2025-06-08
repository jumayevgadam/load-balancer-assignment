package golb

import (
	"container/heap"
	"context"
	"log"
	"net/url"
	"sync"
)

// backendHeap implements a min-heap based on backend load and address.
type backendHeap []*BackendImpl

// Len returns the number of items in the heap.
func (h backendHeap) Len() int {
	return len(h)
}

// Less compares two backends by load, and by address if loads are equal.
func (h backendHeap) Less(i, j int) bool {
	return h[i].GetLoad() < h[j].GetLoad() || (h[i].GetLoad() == h[j].GetLoad() && h[i].addr < h[j].addr)
}

// Swap swaps two items in the heap.
func (h backendHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

// Push adds a new backend to the heap.
func (h *backendHeap) Push(x interface{}) {
	backendImpl, ok := x.(*BackendImpl)
	if !ok {
		log.Println("h.backendHeap.Push: type assertion")
		return
	}

	*h = append(*h, backendImpl)
}

// Pop removes and returns the last backend from the heap.
func (h *backendHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]

	return x
}

// Ensure AdvancedBalancer implements Backend interface.
var _ Backend = (*AdvancedBalancer)(nil)

// AdvancedBalancer implements a heap-based load balancer with health checks.
type AdvancedBalancer struct {
	backends []*BackendImpl // All available backends.
	heap     *backendHeap   // Min-Heap of healthy backend servers.
	mu       sync.Mutex     // Mutex for safe concurrent access.
}

// NewAdvancedLoadBalancer initializes an AdvancedBalancer from a list of backend URLs.
func NewAdvancedLoadBalancer(urls []*url.URL) *AdvancedBalancer {
	backends := make([]*BackendImpl, len(urls))
	h := &backendHeap{}

	for i, u := range urls {
		backends[i] = NewBackend(u.Host)
		*h = append(*h, backends[i])
	}

	heap.Init(h) // initialize a heap structure.

	return &AdvancedBalancer{backends: backends, heap: h}
}

// GetNextServer returns the next available healthy backend based on lowest load.
func (b *AdvancedBalancer) GetNextServer() *BackendImpl {
	b.mu.Lock()
	defer b.mu.Unlock()

	// If heap is empty, repopulate it with healthy backends
	if b.heap.Len() == 0 {
		for _, backend := range b.backends {
			if backend.IsHealthy() {
				heap.Push(b.heap, backend)
			}
		}
	}

	if b.heap.Len() == 0 {
		return nil // no healthy backends available.
	}
	// Select the healthiest backend with lowest load.
	var selected *BackendImpl

	for b.heap.Len() > 0 {
		backend, ok := heap.Pop(b.heap).(*BackendImpl)
		if !ok {
			log.Println("heap.Pop(b.heap).(*BackendImpl): type assertion error")
			continue // skip to next backend in heap.
		}

		if backend.IsHealthy() {
			selected = backend
			break
		}
	}
	// Rebuild heap with all healthy backends.
	for _, backend := range b.backends {
		if backend.IsHealthy() {
			heap.Push(b.heap, backend)
		}
	}

	return selected
}

// Invoke sends the request to the next available backend.
// Marks backend as unhealthy if it fails repeatedly.
func (b *AdvancedBalancer) Invoke(ctx context.Context, req Request) (Response, error) {
	backend := b.GetNextServer()
	if backend == nil {
		return nil, ErrBackendServersEmpty
	}

	resp, err := backend.Invoke(ctx, req)
	if err != nil {
		backend.mu.Lock()
		if backend.failureCount >= maxFailureCount {
			backend.MarkUnhealthy() // Mark backend as unhealthy after maxFailureCount.
		}
		backend.mu.Unlock()
	}

	return resp, err
}
