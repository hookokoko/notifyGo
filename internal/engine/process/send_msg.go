package process

import (
	"context"
	"encoding/json"
	"notifyGo/internal"
	"notifyGo/internal/engine"
)

type MsgSendAction struct {
	producer *engine.Producer
}

func NewMsgSendAction() *MsgSendAction {
	return &MsgSendAction{
		producer: engine.NewProducer([]string{"127.0.0.1:9092"}),
	}
}

func (ms *MsgSendAction) Process(_ context.Context, taskInfo internal.TaskInfo, messageTemplateId int64) error {
	// 发送消息
	taskBytes, err := json.Marshal(taskInfo)
	if err != nil {
		return err
	}

	ms.producer.Send("message_common", taskBytes)
	return nil
}
