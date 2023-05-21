package service

import (
	"context"
	"notifyGo/internal"
)

// 请求target服务，获取发送目标
// 请求content服务，获取发送内容
// 发送到mq

type Core struct {
	ContentService   *ContentService
	TargetService    *TargetService
	SendService      *SendService
	SendBatchService *SendBatchService
}

func NewCore() *Core {
	return &Core{
		ContentService:   NewContentService(),
		TargetService:    NewTargetService(),
		SendService:      NewSendService(),
		SendBatchService: NewSendBatchService(),
	}
}

func (c *Core) Send(ctx context.Context, channel string, target internal.Target, templateId uint64,
	variable map[string]interface{}) error {
	// 获取发送内容
	msgContent := c.ContentService.GetContent(target, templateId, variable)

	task := internal.Task{
		MsgId:       0,
		SendChannel: channel,
		MsgContent:  msgContent,
		MsgReceiver: target,
	}

	return c.SendService.Process(ctx, task)
}

func (c *Core) SentBatch(ctx context.Context, channel string, targetId, templateId uint64,
	variable map[string]interface{}) error {

	// 获取发送目标
	receivers := c.TargetService.GetTarget(targetId)

	batchTask := make([]internal.Task, 0, len(receivers))
	for _, r := range receivers {
		// 针对每一个发送目标 构建发送内容
		// 可以考虑这个加一个缓存
		msgContent := c.ContentService.GetContent(r, templateId, variable)
		batchTask = append(batchTask, internal.Task{
			MsgId:       0,
			SendChannel: channel,
			MsgContent:  msgContent,
			MsgReceiver: r,
		})
	}

	return c.SendBatchService.Process(ctx, batchTask)
}

func (c *Core) SentBox() {

}
