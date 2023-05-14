package engine

import (
	"context"
	"notifyGo/internal"
	"notifyGo/internal/engine/sender"

	"github.com/panjf2000/ants/v2"
)

type TaskRun interface {
	Run(ctx context.Context)
}

type TaskExecutor struct {
	pools map[string]*ants.Pool
}

type Task struct {
	TaskInfo      *internal.TaskInfo
	HandlerAction sender.IHandler
}
