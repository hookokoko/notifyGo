package sender

import (
	"context"
	"fmt"
	"notifyGo/internal"
)

type EmailHandler struct{}

func NewEmailHandler() *EmailHandler {
	return &EmailHandler{}
}

func (eh *EmailHandler) Name() string {
	return internal.EMAIL
}

func (eh *EmailHandler) Execute(ctx context.Context, task *internal.TaskInfo) error {
	fmt.Printf("send email success, %v\n", *task)
	return nil
}
