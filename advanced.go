package golb

import (
	"container/heap"
	"context"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"
)

const healthCheckerTime = 10 * time.Second

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
	backends []*BackendImpl // all unhealthy backends.
	heap     *backendHeap   // Min-Heap of healthy backend servers.
	mu       sync.Mutex     // Mutex for safe concurrent access.
	stopChan chan struct{}  // Channel to stop health checker.
}

// NewAdvancedLoadBalancer initializes an AdvancedBalancer from a list of backend URLs.
func NewAdvancedLoadBalancer(urls []*url.URL) *AdvancedBalancer {
	backends := make([]*BackendImpl, len(urls))
	h := &backendHeap{}
	heap.Init(h)

	for i, u := range urls {
		backends[i] = NewBackend(u.Host)
		// only add healthy backends to heap initially.
		if backends[i].IsHealthy() {
			heap.Push(h, backends[i])
		}
	}

	balancer := &AdvancedBalancer{
		heap:     h,
		backends: backends,
		stopChan: make(chan struct{}),
	}

	// start health checker in separate goroutine.
	go balancer.healthChecker()

	return balancer
}

func (b *AdvancedBalancer) healthChecker() {
	ticker := time.NewTicker(healthCheckerTime) // check every 10 seconds.
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.mu.Lock()
			for _, backend := range b.backends {
				if !backend.IsHealthy() {
					continue // we need to skip if backend unhealthy.
				}

				if backend.IsHealthy() {
					alreadyInHeap := false

					for _, bh := range *b.heap { // bh is backendInHeap and we need to check that backend already stored in heap or not.
						if bh == backend {
							alreadyInHeap = true
							break
						}
					}
					// if not pushed to heap yet, then push healthy backend to heap.
					if !alreadyInHeap {
						heap.Push(b.heap, backend)
					}
				}
			}
			b.mu.Unlock()
		case <-b.stopChan:
			return
		}
	}
}

// StopHealthChecker stops the health checker goroutine.
func (b *AdvancedBalancer) StopHealthChecker() {
	close(b.stopChan)
}

// GetNextServer returns the next available healthy backend based on lowest load.
func (b *AdvancedBalancer) GetNextServer() *BackendImpl {
	b.mu.Lock()
	defer b.mu.Unlock()

	// If heap is empty, repopulate it with healthy backends.
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
		// we remove the least loaded backend from heap.
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

	b.mu.Lock()
	defer b.mu.Unlock()

	if err != nil {
		backend.mu.Lock() // we need to lock backend's own to safely update failure count.
		backend.failureCount++

		if backend.failureCount >= maxFailureCount {
			backend.MarkUnhealthy() // Mark backend as unhealthy after maxFailureCount.
		} else {
			// if not maxFailures yet, push to heap.
			heap.Push(b.heap, backend)
		}
		backend.mu.Unlock()

		return nil, fmt.Errorf("[advanced.backend.Invoke]: failed request: %w", err) // we need to return err for request failed.
	}
	// if backend succeeded, so we return it to heap.
	heap.Push(b.heap, backend)

	return resp, err
}
