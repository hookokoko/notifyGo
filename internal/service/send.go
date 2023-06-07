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
	cfg := mq.NewConfig("/Users/hooko/GolandProjects/notifyGo/config/kafka_topic.toml")
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

	msgType := task.MsgContent.Type
	channel := task.SendChannel

	ss.producer.Send(channel, msgType, taskBytes)
	return nil
}
