package golb

import "net/url"

func NewBalancer(level string, urls []*url.URL) Backend {
	switch level {
	case "basic":
		return NewBasicLoadBalancer(urls)
	case "intermediate":
		return NewIntermediateLoadBalancer(urls)
	case "advanced":
		return NewAdvancedLoadBalancer(urls)
	}

	return nil
}
