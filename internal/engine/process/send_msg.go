package process

import "context"

type MsgSendAction struct {
}

func NewMsgSendAction() *MsgSendAction {
	return &MsgSendAction{}
}

func (p *MsgSendAction) Process(_ context.Context, msgInfo any, messageTemplateId int64) error {
	// 发送消息装配
	return nil
}
