package clientX

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
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

	// TODO 鉴权
	logRecord.PointStart("conn_cost")
	conn, _ := grpc.DialContext(ctx, to.GetGrpcReqDomain(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(&GrpcStat{L: logRecord}),
		grpc.WithBlock(),
	)
	logRecord.PointStop("conn_cost")

	in := g.grpcReq.Data
	out := g.grpcResp

	logRecord.Method = g.grpcReq.Method
	fullMethodName := fmt.Sprintf("/%s.%s/%s", g.grpcReq.Package, g.grpcReq.Service, g.grpcReq.Method)
	logRecord.Path = fullMethodName

	//logRecord.PointStart("rpc_cost") // 这种好像也一样
	err := conn.Invoke(ctx, fullMethodName, in, out)
	//logRecord.PointStop("rpc_cost")

	s, _ := status.FromError(err)
	logRecord.AddField("status", s.Code())

	if err != nil {
		return err
	}

	return nil
}

func (g *GrpcProtocol) Result() any {
	return g.grpcResp
}

type GrpcStat struct {
	L *LogRecord
}

func (g *GrpcStat) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	return ctx
}

func (g *GrpcStat) HandleRPC(ctx context.Context, rpcStats stats.RPCStats) {
	switch s := rpcStats.(type) {
	case *stats.Begin:
	case *stats.End:
		g.L.AddTimeCostPoint("rpc_cost", s.EndTime.Sub(s.BeginTime))
	}
}

func (g *GrpcStat) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	return ctx
}

func (g *GrpcStat) HandleConn(ctx context.Context, connStats stats.ConnStats) {}
