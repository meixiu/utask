// Copyright 2019 Yoozoo Authors. All Rights Reserved.
// @Description:monitor interface for producer and consumer

package monitor

import (
	"time"

	"github.com/meixiu/utask/task"
)

// ConsumerMonitor consumer monitor
type ConsumerMonitor interface {
	// PullTask pull task from taskStore
	PullTask(cid string, task task.Tasker)
	// HandleTask handle task
	HandleTask(cid string, task task.Tasker)
	// Retries retries task
	Retries(cid string, task task.Tasker)
}

// ProducerMonitor producer monitorF
type ProducerMonitor interface {
	Request(processType string)
	Latency(method, path, processType string, duration time.Duration)
}
