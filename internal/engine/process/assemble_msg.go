package process

import (
	"context"
	"notifyGo/internal"
)

type AssembleParamAction struct {
}

func NewAssembleParamAction() *AssembleParamAction {
	return &AssembleParamAction{}
}

func (p *AssembleParamAction) Process(_ context.Context, taskInfo internal.TaskInfo, messageTemplateId int64) error {
	// 发送消息装配
	return nil
}
