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
	return internal.SMSNAME
}

func (fh *SmsHandler) Execute(ctx context.Context, task *internal.Task) error {
	return nil
}
