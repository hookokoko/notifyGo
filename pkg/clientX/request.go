package clientX

import (
	"context"
	"fmt"
	"log"

	"github.com/go-resty/resty/v2"
)

type Protocol interface {
	Do(ctx context.Context, to *Addr) error
	Result() any
}

type HttpProtocol struct {
	restyClient   *resty.Client
	restyResponse *resty.Response
	originRequest *HttpRequest
	result        any
	service       Service
	isHTTP        bool
}

type HttpRequest struct {
	Header      map[string]string
	Method      string
	Body        map[string]any
	Path        string
	QueryParams map[string]string
}

func NewHttpProtocol(request *HttpRequest, result any, isHttp bool) *HttpProtocol {
	client := resty.New()

	return &HttpProtocol{
		restyClient:   client,
		result:        result,
		originRequest: request,
		isHTTP:        isHttp,
	}
}

func (h *HttpProtocol) Do(ctx context.Context, to *Addr) error {
	h.restyClient.SetBaseURL(to.GetReqDomain(h.isHTTP) + h.originRequest.Path)

	rr := h.restyClient.NewRequest().
		SetContext(ctx).
		SetHeaders(h.originRequest.Header).
		SetBody(h.originRequest.Body).
		SetQueryParams(h.originRequest.QueryParams).
		SetResult(h.result)
	rr.Method = h.originRequest.Method

	resp, err := rr.Send()

	h.restyResponse = resp
	return err
}

func (h *HttpProtocol) Result() any {
	return h.result
}

func Go(ctx context.Context, srvName string, request any, result any) error {
	// 根据srvName获取请求类型
	srvI, ok := ServicesMap.Load(srvName)
	if !ok {
		log.Fatalf("找不到配置的服务: %s", srvName)
	}
	srv, ok := srvI.(*Service)
	if !ok {
		return fmt.Errorf("*Service类型断言错误%s\n", srvName)
	}
	// 根据负载均衡策略获取请求的目标
	to := srv.PickTarget()

	var hp Protocol
	switch typ := request.(type) {
	case *HttpRequest:
		proto := srv.Protocol
		if proto == "https" {
			hp = NewHttpProtocol(typ, result, false)
		} else if proto == "http" {
			hp = NewHttpProtocol(typ, result, true)
		} else {
			return fmt.Errorf("服务协议和请求协议不一致, %s", proto)
		}
	default:
		return fmt.Errorf("unsupport request type, %s", typ)
	}

	err := hp.Do(ctx, to)
	if err != nil {
		return err
	}

	return nil
}
