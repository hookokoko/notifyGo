package process

import "context"

type AssembleParamAction struct {
}

func NewAssembleParamAction() *AssembleParamAction {
	return &AssembleParamAction{}
}

func (p *AssembleParamAction) Process(_ context.Context, msgInfo any, messageTemplateId int64) error {
	// 发送消息装配
	return nil
}
