// Copyright 2019 Yoozoo Authors. All Rights Reserved.
// @Description:monitor interface for producer and consumer

package monitor

import (
	"time"
	"utask/task"
)

// ConsumerMonitor consumer monitor
type ConsumerMonitor interface {
	// PullTask pull task from taskStore
	PullTask(cid string, task task.Tasker)
	// HandleTask handle task
	HandleTask(cid string, duration time.Duration, task task.Tasker)
	// Retries retries task
	Retries(cid string, task task.Tasker)
}

// ProducerMonitor producer monitorF
type ProducerMonitor interface {
	Request(appID, processType string)
	Latency(appID, processType string, duration time.Duration)
}
