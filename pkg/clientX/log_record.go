package clientX

import (
	"notifyGo/pkg/logger"
	"sync"
	"time"
)

// 日志record应该具备如下功能：
// 1. 记录某个统计项的耗时
// 2. 记录时间戳
// 3. 日志落盘

type LogRecord struct {
	RspCode             int
	Protocol            string
	Method              string
	Path                string
	IDC                 string
	IPPort              string
	Host                string
	Error               error
	retry               int
	timeCostStatics     map[string]*StaticsItem
	timeCostStaticsLock sync.Mutex
	timeCostPoints      map[string]time.Duration
	timeCostPointsLock  sync.Mutex
	fields              map[string]any
	fieldsLock          sync.Mutex
}

type StaticsItem struct {
	StartPoint time.Time
	StopPoint  time.Time
}

func (s *StaticsItem) GetSpan() time.Duration {
	return s.StopPoint.Sub(s.StartPoint)
}

func NewLogRecord() *LogRecord {
	return &LogRecord{
		timeCostStatics: make(map[string]*StaticsItem),
		timeCostPoints:  make(map[string]time.Duration),
		fields:          make(map[string]any),
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
	item, ok := lr.timeCostStatics[name]
	if ok {
		item.StopPoint = time.Now()
	}
}

func (lr *LogRecord) AddField(name string, value any) {
	defer lr.fieldsLock.Unlock()
	lr.fieldsLock.Lock()
	lr.fields[name] = value
}

func (lr *LogRecord) AddTimeCostPoint(name string, d time.Duration) {
	defer lr.timeCostPointsLock.Unlock()
	lr.timeCostPointsLock.Lock()
	lr.timeCostPoints[name] = d
}

func (lr *LogRecord) Flush() {
	// 这里面有个两个问题
	// 1. 日志的级别如何定义
	// 2. 如何将日志字段格式话输出
	// 3. 获取trace的字段
	// 基本功能实现了，总感觉不是很优雅
	field := make([]logger.Field, 0, len(lr.timeCostStatics)+7)
	field = append(field, logger.Int("code", lr.RspCode))
	field = append(field, logger.String("path", lr.Path))
	field = append(field, logger.String("idc", lr.IDC))
	field = append(field, logger.String("ipport", lr.IPPort))
	field = append(field, logger.String("host", lr.Host))
	field = append(field, logger.Int("retry", lr.retry))
	field = append(field, logger.String("protocol", lr.Protocol))
	field = append(field, logger.String("method", lr.Method))

	for name, sItem := range lr.timeCostStatics {
		span := sItem.GetSpan()
		field = append(field, logger.Duration(name, span))
	}

	for name, f := range lr.fields {
		field = append(field, logger.Any(name, f))
	}

	// TODO，ms时间要保留几位小数, 当前打印的日志是只展示整数部分
	for name, d := range lr.timeCostPoints {
		field = append(field, logger.Duration(name, d))
	}

	if lr.Error != nil {
		logger.Default().Error(lr.Error.Error(), field...)
	} else {
		logger.Default().Info("", field...)
	}
}
