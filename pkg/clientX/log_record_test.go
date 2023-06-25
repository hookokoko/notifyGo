package clientX

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func Test_Example(t *testing.T) {
	le := NewLogRecord()
	le.RspCode = http.StatusOK
	le.Path = "/home/1"
	le.IDC = "jja"
	le.retry = 1
	le.Host = "127.0.0.1"
	le.IPPort = "127.0.0.1:8080"
	le.Error = fmt.Errorf("错了")

	le.PointStart("connection")
	time.Sleep(20 * time.Millisecond)
	le.PointStop("connection")

	le.PointStart("cost")
	time.Sleep(100 * time.Millisecond)
	le.PointStop("cost")

	le.Flush()
}
