package service

import (
	"context"
	"encoding/json"
	"notifyGo/internal"
	"notifyGo/pkg/mq"
)

type SendService struct {
	producer *mq.Producer
}

func NewSendService() *SendService {
	cfg := mq.NewConfig([]string{"127.0.0.1:9092"})
	return &SendService{
		producer: mq.NewProducer(cfg),
	}
}

func (ss *SendService) Process(_ context.Context, task internal.Task) error {
	// 发送消息
	taskBytes, err := json.Marshal(task)
	if err != nil {
		return err
	}

	// topic的选择要抽象出来
	ss.producer.Send("email", taskBytes)
	return nil
}
