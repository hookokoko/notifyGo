package sender

import (
	"context"
	"notifyGo/internal"
)

type FeishuHandler struct{}

func NewFeishuHandler() *FeishuHandler {
	return &FeishuHandler{}
}

func (fh *FeishuHandler) Name() string {
	return internal.FEISHU
}

func (fh *FeishuHandler) Execute(ctx context.Context, task *internal.Task) error {
	return nil
}
