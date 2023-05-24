package service

import (
	"context"
	"encoding/json"
	"notifyGo/internal"
	"notifyGo/pkg/mq"
)

type SendBatchService struct {
	producer *mq.Producer
}

func NewSendBatchService() *SendBatchService {
	cfg := mq.NewConfig([]string{"127.0.0.1:9092"})
	return &SendBatchService{
		producer: mq.NewProducer(cfg),
	}
}

func (ss *SendBatchService) Process(_ context.Context, tasks []internal.Task) error {
	// 发送消息
	// 这里如果是100w条消息，内存会炸吧？
	// 所以这里进行流式接收、流式处理
	// 这里还会写入db
	for _, task := range tasks {
		taskBytes, err := json.Marshal(task)
		if err != nil {
			return err
		}

		// topic的选择要抽象出来
		ss.producer.Send("email", taskBytes)
	}

	return nil
}
