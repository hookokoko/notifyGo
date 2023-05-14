package engine

import (
	"context"
	"log"
	"notifyGo/internal"
	"notifyGo/internal/engine/sender"
	"runtime"

	"github.com/panjf2000/ants/v2"
)

// NewTaskExecutor 带协程池的task执行器
func NewTaskExecutor() *TaskExecutor {
	pools := make(map[string]*ants.Pool)
	defaultPool, err := ants.NewPool(runtime.NumCPU())
	if err != nil {
		log.Fatalf("创建ants pool出错，%v", err)
	}
	pools["default"] = defaultPool
	return &TaskExecutor{
		pools: pools,
	}
}

// Submit 把任务提交到对应的池子内
func (t *TaskExecutor) Submit(ctx context.Context, groupId string, run TaskRun) error {
	pool, ok := t.pools[groupId]
	if !ok {
		pool = t.pools["default"]
	}
	return pool.Submit(func() {
		run.Run(ctx)
	})
}

// NewTask 具体的一个要执行的Task, 当前主要待执行的任务信息TaskInfo
func NewTask(taskInfo *internal.TaskInfo, ha sender.IHandler) *Task {
	return &Task{
		TaskInfo:      taskInfo,
		HandlerAction: ha,
	}
}

func (t *Task) Run(ctx context.Context) {
	// 执行对应类型的发送, 这里的context最好通过消息体中的字段重新构造一下
	_ = t.HandlerAction.Execute(ctx, t.TaskInfo)
	return
}
