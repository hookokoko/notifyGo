package service

import (
	"context"
	"fmt"
	"notifyGo/internal"
	"notifyGo/internal/model"

	"golang.org/x/sync/errgroup"
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
func (c *Core) Send(ctx context.Context, channel string, msgType string, target internal.ITarget, templateId int64,
	variable map[string]interface{}) error {
	// 获取发送内容
	msg := c.ContentService.GetContent(target, templateId, variable)
	msg.Type = msgType

	err := c.NotifyGoDAO.InsertRecord(ctx, templateId, target, msg.Content)
	if err != nil {
		return err
	}

	task := internal.Task{
		//TaskId:       delivery.Id, // 这里考虑手动生成，现在先不传
		SendChannel: channel,
		MsgContent:  msg,
		MsgReceiver: target,
	}

	return c.SendService.Process(ctx, task)
}

// 1. 创建一个delivery记录
// 2. 获取所有的target，并创建target记录关联delivery id
// 3. 推送至kafka。批量的话如何推？如何做到一边获取target一边发送，即流式推送
func (c *Core) SentBatch(ctx context.Context, channel string, targetId, templateId int64,
	variable map[string]interface{}) error {

	// 获取发送目标, 这个相当于一个服务，支持分页获取
	// 或者流式获取发送目标，流式构造发送内容
	receivers := c.TargetService.GetTarget(targetId)

	batchTask := make([]internal.Task, 0, len(receivers))

	var eg errgroup.Group
	for _, recvr := range receivers {
		// 针对每一个发送目标 构建发送内容
		// 可以考虑这个加一个缓存
		// 还应该使用到goroutine池，并发获取并发送
		// 获取内容后修改target表状态为 待发送 ，并写入content
		eg.Go(func() error {
			msgContent := c.ContentService.GetContent(recvr, templateId, variable)
			batchTask = append(batchTask, internal.Task{
				//MsgId:       0,
				SendChannel: channel,
				MsgContent:  msgContent,
				MsgReceiver: recvr,
			})
			// 这里的error可以加工一下带上这次执行的标识信息
			return c.SendBatchService.Process(ctx, batchTask)
		})
	}
	if err := eg.Wait(); err != nil {
		fmt.Printf("get error:%v\n", err)
	}
	return nil
}

func (c *Core) SentBox() {

}
