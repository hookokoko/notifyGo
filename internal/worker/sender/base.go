package sender

import (
	"context"
	"notifyGo/internal"
	"notifyGo/pkg/item"
)

type IHandler interface {
	Name() string
	Execute(ctx context.Context, taskInfo *internal.Task) error
	//Allow(ctx context.Context, taskInfo *TaskInfo) bool
}

type HandleManager struct {
	manager *item.Manager
}

func NewHandlerManager() *HandleManager {
	return &HandleManager{
		manager: item.NewManager(
			NewEmailHandler(),
			NewSmsHandler(),
			NewPushHandler(),
		),
	}
}

func (hm *HandleManager) Get(key string) (resp IHandler, err error) {
	if h, err := hm.manager.Get(key); err == nil {
		return h.(IHandler), nil
	}
	return nil, err
}
