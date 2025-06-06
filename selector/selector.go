// This file keeps common types.
package selector

import (
	"context"
	"errors"
	"fmt"

	"github.com/jumayevgadam/golb"
)

var (
	_ golb.Backend = (*BasicBalancer)(nil)
	_ golb.Backend = (*backendImpl)(nil)
)

// ErrBackendServersEmpty is returned when no backends are available.
var ErrBackendServersEmpty = errors.New("backend server list is empty")

// backendImpl represents a backend server.
type backendImpl struct {
	addr string
}

// Invoke processes a request on the backend.
func (b *backendImpl) Invoke(ctx context.Context, req golb.Request) (golb.Response, error) {
	return fmt.Sprintf("addr: %s, req: %v", b.addr, req), nil
}

// Host returns the backend's address.
func (b *backendImpl) Host() string {
	return b.addr
}
