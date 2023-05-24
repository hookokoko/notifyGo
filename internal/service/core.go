package service

import (
	"context"
	"notifyGo/internal"
	"notifyGo/internal/model"
	"time"
)

// 请求target服务，获取发送目标
// 请求content服务，获取发送内容
// 发送到mq

type Core struct {
	ContentService   *ContentService
	TargetService    *TargetService
	SendService      *SendService
	SendBatchService *SendBatchService
	NotifyGoDAO      *model.NotifyGoDAO
}

func NewCore() *Core {
	return &Core{
		ContentService:   NewContentService(),
		TargetService:    NewTargetService(),
		SendService:      NewSendService(),
		SendBatchService: NewSendBatchService(),
		// 这里是否需要自己管理连接池
		NotifyGoDAO: model.NewNotifyGoDAO(),
	}
}

// 1. 创建一个delivery记录
// 2. 获取所有的target，并创建target记录关联delivery id
// 3. 推送至kafka
func (c *Core) Send(ctx context.Context, channel string, target internal.Target, templateId int64,
	variable map[string]interface{}) error {
	// 获取发送内容
	msgContent := c.ContentService.GetContent(target, templateId, variable)

	// 在DAO层封装事务，target和delivery表
	delivery := model.Delivery{
		TemplateId:  templateId,
		Status:      1, // 消息创建状态
		SendChannel: 40,
		MsgType:     20,
		Proposer:    "crm",
		Creator:     "chenhaokun",
		Updator:     "chenhaokun",
		IsDelted:    0,
		Created:     time.Now(),
		Updated:     time.Now(),
	}

	// 这里如何获取插入的id
	_, err := c.NotifyGoDAO.Insert(&delivery)
	if err != nil {
		return err
	}

	tgt := model.Target{
		TargetIdType: 10, // 邮箱，这里封装一个枚举，需要根据 name(target.Email) -> value
		TargetId:     target.Email,
		DeliveryId:   delivery.Id,
		Status:       1, // 创建状态
		MsgContent:   msgContent,
	}
	_, err = c.NotifyGoDAO.InsertOne(&tgt)
	if err != nil {
		return err
	}

	task := internal.Task{
		MsgId:       delivery.Id,
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
