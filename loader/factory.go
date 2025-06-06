package loader

import (
	"net/url"

	"github.com/jumayevgadam/golb"
	"github.com/jumayevgadam/golb/selector"
)

func NewBalancer(level string, urls []*url.URL) golb.Backend {
	switch level {
	case "basic":
		return selector.NewBasicLoadBalancer(urls)
	}

	return nil
}
