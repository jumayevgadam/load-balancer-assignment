package golb

import "errors"

// ErrBackendServersEmpty is returned when backend list is empty.
var ErrBackendServersEmpty = errors.New("backend server list is empty")

// ErrNoAvailableBackends is returned when no backends are available.
var ErrNoAvailableBackends = errors.New("no available backends")
