package clientX

import (
	"sync"
	"time"
)

// 日志record应该具备如下功能：
// 1. 记录某个统计项的耗时
// 2. 记录时间戳
// 3. 日志落盘

type LogRecord struct {
	RspCode             int
	Path                string
	IDC                 string
	IPPort              string
	Host                string
	Error               error
	index               int
	timeCostStatics     map[string]*StaticsItem
	timeCostStaticsLock sync.Mutex
	timePoints          map[string]time.Time
	timePointsLock      sync.Mutex
}

type StaticsItem struct {
	StartPoint time.Time
	StopPoint  time.Time
}

func NewLogRecord() *LogRecord {
	return &LogRecord{
		timeCostStatics: make(map[string]*StaticsItem),
		timePoints:      make(map[string]time.Time),
	}
}

func (lr *LogRecord) PointStart(name string) {
	defer lr.timeCostStaticsLock.Unlock()
	lr.timeCostStaticsLock.Lock()
	item := new(StaticsItem)
	item.StartPoint = time.Now()
	lr.timeCostStatics[name] = item
}

func (lr *LogRecord) PointStop(name string) {
	defer lr.timeCostStaticsLock.Unlock()
	lr.timeCostStaticsLock.Lock()
	item := new(StaticsItem)
	item.StopPoint = time.Now()
	lr.timeCostStatics[name] = item
}

func (lr *LogRecord) Point(name string) {
	defer lr.timePointsLock.Unlock()
	lr.timePointsLock.Lock()
	lr.timePoints[name] = time.Now()
}

func (lr *LogRecord) Flush() {}
