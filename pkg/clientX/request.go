package clientX

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"
)

func Go(ctx context.Context, srvName string, request any, result any) error {
	// 塞到ctx里，但是不是很优雅？
	l := NewLogRecord()
	l.PointStart("total")
	defer l.Flush()
	defer l.PointStop("total")

	// 根据srvName获取请求类型
	srvI, ok := ServicesMap.Load(srvName)
	if !ok {
		return fmt.Errorf("找不到配置的服务: %s", srvName)
	}
	srv, ok := srvI.(*Service)
	if !ok {
		return fmt.Errorf("*Service类型断言错误%s\n", srvName)
	}

	// 根据负载均衡策略获取请求的目标
	l.PointStart("get_target_cost")
	to := srv.PickTarget()
	if to == nil {
		return fmt.Errorf("获取请求的目标为空%s\n", srvName)
	}
	l.PointStop("get_target_cost")

	l.Host = to.Host
	l.IPPort = to.IP + ":" + to.Port
	l.IDC = to.IDC

	p := srv.Protocol
	l.Protocol = p

	var hp Protocol
	switch req := request.(type) {
	case *HttpRequest:
		if p == "https" {
			hp = NewHttpProtocol(req, result, false)
		} else if p == "http" {
			hp = NewHttpProtocol(req, result, true)
		} else {
			return fmt.Errorf("服务协议和请求协议不一致, %s", p)
		}
	case *GrpcRequest:
		hp = NewGrpcProtocol(req, result.(proto.Message))
	default:
		return fmt.Errorf("不支持的请求类型, %s", req)
	}

	valueCtx := context.WithValue(ctx, "logRecord", l)
	err := hp.Do(valueCtx, to)
	if err != nil {
		return err
	}

	return nil
}
