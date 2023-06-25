package clientX

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"log"
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
	client := resty.New().EnableTrace()

	return &HttpProtocol{
		restyClient:   client,
		result:        result,
		originRequest: request,
		isHTTP:        isHttp,
	}
}

func (h *HttpProtocol) Do(ctx context.Context, to *Addr) error {
	logRecordI := ctx.Value("logRecord")
	if logRecordI == nil {
	}
	logRecord, ok := logRecordI.(*LogRecord)
	if !ok {
	}

	h.restyClient.SetBaseURL(to.GetReqDomain(h.isHTTP) + h.originRequest.Path)

	rr := h.restyClient.NewRequest().
		SetContext(ctx).
		SetHeaders(h.originRequest.Header).
		SetBody(h.originRequest.Body).
		SetQueryParams(h.originRequest.QueryParams).
		SetResult(h.result)

	rr.Method = h.originRequest.Method
	logRecord.Method = rr.Method
	logRecord.Path = h.originRequest.Path

	resp, err := rr.Send()
	logRecord.RspCode = resp.StatusCode()
	if err != nil {
		logRecord.Error = err
	}

	// set trace info
	ti := resp.Request.TraceInfo()
	logRecord.AddTimeCostPoint("net_cost", ti.TotalTime)
	logRecord.AddTimeCostPoint("connect_cost", ti.ConnTime)
	logRecord.AddTimeCostPoint("dns_cost", ti.DNSLookup)
	logRecord.AddTimeCostPoint("server_cost", ti.ServerTime)
	logRecord.AddTimeCostPoint("resp_cost", ti.ResponseTime)
	logRecord.AddTimeCostPoint("tcp_cost", ti.TCPConnTime)
	logRecord.AddTimeCostPoint("tls_cost", ti.TLSHandshake)

	h.restyResponse = resp
	return err
}

func (h *HttpProtocol) Result() any {
	return h.result
}

func Go(ctx context.Context, srvName string, request any, result any) error {
	// newLogRecord, 将record塞到context里？
	l := NewLogRecord()
	defer l.Flush()

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
		return fmt.Errorf("unsupport request type, %s", typ)
	}

	valueCtx := context.WithValue(ctx, "logRecord", l)
	err := hp.Do(valueCtx, to)
	if err != nil {
		return err
	}

	return nil
}
