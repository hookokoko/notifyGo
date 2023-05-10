package process

import "context"

type ParamCheckAction struct {
}

func NewPreParamCheckAction() *ParamCheckAction {
	return &ParamCheckAction{}
}

func (p *ParamCheckAction) Process(_ context.Context, msgInfo any, messageTemplateId int64) error {
	// 是否包含收件人检查
	// 模版是否存在检查
	return nil
}
