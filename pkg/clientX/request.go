package clientX

import (
	"context"
	"fmt"
)

func Go(ctx context.Context, srvName string, request any, result any) error {
	// 塞到ctx里，但是不是很优雅？
	l := NewLogRecord()
	defer l.Flush()

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

	proto := srv.Protocol
	l.Protocol = proto

	var hp Protocol
	switch typ := request.(type) {
	case *HttpRequest:
		if proto == "https" {
			hp = NewHttpProtocol(typ, result, false)
		} else if proto == "http" {
			hp = NewHttpProtocol(typ, result, true)
		} else {
			return fmt.Errorf("服务协议和请求协议不一致, %s", proto)
		}
	default:
		return fmt.Errorf("不支持的请求类型, %s", typ)
	}

	valueCtx := context.WithValue(ctx, "logRecord", l)
	err := hp.Do(valueCtx, to)
	if err != nil {
		return err
	}

	return nil
}
