package process

import "context"

type Process interface {
	Process(ctx context.Context, msgInfo any, messageTemplateId int64) error
}

type MsgSendProcess struct {
	process []Process
}

func NewMsgSendProcess(pc *ParamCheckAction, ap *AssembleParamAction, ms *MsgSendAction) *MsgSendProcess {
	return &MsgSendProcess{[]Process{pc, ap, ms}}
}

func (m *MsgSendProcess) Process(ctx context.Context, msgInfo any, messageTemplateId int64) error {
	for _, pr := range m.process {
		err := pr.Process(ctx, msgInfo, messageTemplateId)
		if err != nil {
			return err
		}
	}
	return nil
}
