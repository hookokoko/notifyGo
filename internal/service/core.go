package service

import (
	"context"
	"notifyGo/internal"
	"notifyGo/internal/model"
)

// 请求target服务，获取发送目标
// 请求content服务，获取发送内容
// 发送到mq

type Core struct {
	ContentService   *ContentService
	TargetService    *TargetService
	SendService      *SendService
	SendBatchService *SendBatchService
	NotifyGoDAO      model.INotifyGoDAO
}

func NewCore() *Core {
	return &Core{
		ContentService:   NewContentService(),
		TargetService:    NewTargetService(),
		SendService:      NewSendService(),
		SendBatchService: NewSendBatchService(),
		// 这里是否需要自己管理连接池
		NotifyGoDAO: model.NewINotifyGoDAO(),
	}
}

// 1. 创建一个delivery记录
// 2. 创建一个target，并关联delivery id
// 3. 推送至kafka
func (c *Core) Send(ctx context.Context, channel string, target internal.ITarget, templateId int64,
	variable map[string]interface{}) error {
	// 获取发送内容
	msgContent := c.ContentService.GetContent(target, templateId, variable)

	err := c.NotifyGoDAO.InsertRecord(ctx, templateId, target, msgContent)
	if err != nil {
		return err
	}

	task := internal.Task{
		//MsgId:       delivery.Id, // 这里考虑手动生成，现在先不传
		SendChannel: channel,
		MsgContent:  msgContent,
		MsgReceiver: target,
	}

	return c.SendService.Process(ctx, task)
}

// 1. 创建一个delivery记录
// 2. 获取所有的target，并创建target记录关联delivery id
// 3. 推送至kafka。批量的话如何推？如何做到一边获取target一边发送，即流式推送
func (c *Core) SentBatch(ctx context.Context, channel string, targetId, templateId int64,
	variable map[string]interface{}) error {

	// 获取发送目标
	receivers := c.TargetService.GetTarget(targetId)

	batchTask := make([]internal.Task, 0, len(receivers))
	for _, r := range receivers {
		// 针对每一个发送目标 构建发送内容
		// 可以考虑这个加一个缓存
		// 还应该使用到goroutine池，并发获取并发送
		// 获取内容后修改target表状态为 待发送 ，并写入content
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
