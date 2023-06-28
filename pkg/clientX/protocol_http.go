package clientX

import (
	"context"

	"github.com/go-resty/resty/v2"
)

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

	h.restyClient.SetBaseURL(to.GetHttpReqDomain(h.isHTTP) + h.originRequest.Path)

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
