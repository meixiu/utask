package server

import (
	"context"

	"github.com/meixiu/utask/store"
	"github.com/meixiu/utask/task"
)

// Producer producer for add task
type Producer interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

// Push push task
func Push(sid string, store store.TaskStorer, task task.Tasker) (bool, error) {
	task.Init(sid)
	return store.RPush(task)
}
