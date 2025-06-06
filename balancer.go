package golb

import "context"

type (
	Request  interface{}
	Response interface{}
)

type Backend interface {
	Invoke(ctx context.Context, req Request) (Response, error)
}
