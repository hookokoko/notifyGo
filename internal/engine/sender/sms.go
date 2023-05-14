package sender

import (
	"context"
	"notifyGo/internal"
)

type SmsHandler struct{}

func NewSmsHandler() *SmsHandler {
	return &SmsHandler{}
}

func (fh *SmsHandler) Name() string {
	return internal.SMS
}

func (fh *SmsHandler) Execute(ctx context.Context, task *internal.TaskInfo) error {
	return nil
}
