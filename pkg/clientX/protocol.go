package clientX

import (
	"context"
)

type Protocol interface {
	Do(ctx context.Context, to *Addr) error
	Result() any
}
