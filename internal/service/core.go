package service

import (
	"context"
	"notifyGo/internal"
)

// 请求target服务，获取发送目标
// 请求content服务，获取发送内容
// 发送到mq

type Core struct {
	ContentService *ContentService
	TargetService  *TargetService
	SendService    *SendService
}

func NewCore() *Core {
	return &Core{
		ContentService: NewContentService(),
		TargetService:  NewTargetService(),
		SendService:    NewSendService(),
	}
}

func (c *Core) Send(ctx context.Context, channel string, targetId, templateId uint64,
	variable map[string]interface{}) error {
	// 获取发送内容
	msgContent := c.ContentService.GetContent()

	// 获取发送目标
	receivers := c.TargetService.GetTarget()

	task := internal.Task{
		MsgId:       0,
		SendChannel: channel,
		MsgContent:  msgContent,
		MsgReceiver: receivers,
	}

	return c.SendService.Process(ctx, task)
}

func (c *Core) SentBatch() {

}
