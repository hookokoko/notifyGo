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

func NewSendService(mqCfg *mq.Config) *SendService {
	return &SendService{
		producer: mq.NewProducer(mqCfg),
	}
}

func (ss *SendService) Process(_ context.Context, task internal.Task) error {
	// 发送消息
	taskBytes, err := json.Marshal(task)
	if err != nil {
		return err
	}

	ss.producer.Send(task.SendChannel, taskBytes)
	return nil
}
