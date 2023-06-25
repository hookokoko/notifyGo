package clientX

import "context"

type GrpcRequest struct{}

type GrpcProtocol struct{}

func (g *GrpcProtocol) Do(ctx context.Context, to *Addr) error {
	return nil
}

func (g *GrpcProtocol) Result() any {
	return nil
}
