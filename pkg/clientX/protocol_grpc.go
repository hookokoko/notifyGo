package clientX

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcRequest struct {
	Data    proto.Message
	Header  map[string]string
	Package string
	Method  string
	Service string
}

type GrpcProtocol struct {
	grpcReq  *GrpcRequest
	grpcResp proto.Message
}

func NewGrpcProtocol(req *GrpcRequest, res proto.Message) *GrpcProtocol {
	return &GrpcProtocol{
		grpcReq:  req,
		grpcResp: res,
	}
}

func (g *GrpcProtocol) Do(ctx context.Context, to *Addr) error {
	logRecordI := ctx.Value("logRecord")
	if logRecordI == nil {
	}
	logRecord, ok := logRecordI.(*LogRecord)
	if !ok {
	}

	conn, _ := grpc.Dial(to.GetGrpcReqDomain(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	in := g.grpcReq.Data
	out := g.grpcResp

	logRecord.Method = g.grpcReq.Method
	fullMethodName := fmt.Sprintf("/%s.%s/%s", g.grpcReq.Package, g.grpcReq.Service, g.grpcReq.Method)
	logRecord.Path = fullMethodName

	err := conn.Invoke(ctx, fullMethodName, in, out)
	if err != nil {
		return err
	}

	return nil
}

func (g *GrpcProtocol) Result() any {
	return g.grpcResp
}
