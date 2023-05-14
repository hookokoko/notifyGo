package process

import (
	"context"
	"notifyGo/internal"
)

type Process interface {
	Process(ctx context.Context, taskInfo internal.TaskInfo, messageTemplateId int64) error
}

type MsgSendProcess struct {
	process []Process
}

func NewMsgSendProcess() *MsgSendProcess {
	return &MsgSendProcess{
		[]Process{
			NewPreParamCheckAction(),
			NewAssembleParamAction(),
			NewMsgSendAction(),
		},
	}
}

func (m *MsgSendProcess) Process(ctx context.Context, taskInfo internal.TaskInfo, messageTemplateId int64) error {
	for _, pr := range m.process {
		err := pr.Process(ctx, taskInfo, messageTemplateId)
		if err != nil {
			return err
		}
	}
	return nil
}
