package worker

import (
	"context"
	"log"
	"notifyGo/internal"
	"notifyGo/internal/worker/sender"
	"runtime"

	"github.com/panjf2000/ants/v2"
)

// NewTaskExecutor 带协程池的task执行器
func NewPoolExecutor() *PoolExecutor {
	defaultPool, err := ants.NewPool(runtime.NumCPU())
	if err != nil {
		log.Fatalf("创建ants pool出错，%v", err)
	}
	return &PoolExecutor{
		pool: defaultPool,
	}
}

type PoolExecutor struct {
	pool *ants.Pool
}

// Submit 把任务提交到对应的池子内
func (pe *PoolExecutor) Submit(ctx context.Context, run TaskRun) error {
	return pe.pool.Submit(func() {
		run.Run(ctx)
	})
}

type Task struct {
	Task          *internal.Task
	HandlerAction sender.IHandler
}

// NewTask 具体的一个要执行的Task, 当前主要待执行的任务信息TaskInfo
func NewTask(taskInfo *internal.Task, ha sender.IHandler) *Task {
	return &Task{
		Task:          taskInfo,
		HandlerAction: ha,
	}
}

func (t *Task) Run(ctx context.Context) {
	// 执行对应类型的发送, 这里的context最好通过消息体中的字段重新构造一下
	_ = t.HandlerAction.Execute(ctx, t.Task)
	return
}

type TaskRun interface {
	Run(ctx context.Context)
}
