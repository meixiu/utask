// Copyright 2019 Yoozoo Authors. All Rights Reserved.
// @Description: Monitor storer

package monitor

import (
	"strconv"
	"time"

	"github.com/meixiu/utask/task"

	"github.com/prometheus/client_golang/prometheus"
)

// PromMonitor Consumer prometheus
type PromMonitor struct {
	// PullTaskCounterVec pull task counters
	PullTaskCounterVec *prometheus.CounterVec
	// HandleTaskCounterVec handle task
	HandleTaskCounterVec *prometheus.CounterVec
	// HandleTaskDuration handle task duration
	HandleTaskDuration *prometheus.GaugeVec
	// RetryTaskCounterVec retry
	RetryTaskCounterVec *prometheus.CounterVec
	// RequestCounterVer producer request num
	RequestCounterVer *prometheus.CounterVec
	// ResponseLatency producer latency
	ResponseLatency *prometheus.GaugeVec
}

var (
	// DefaultPromMonitor redis store
	DefaultPromMonitor = NewPromStore()
)

// NewPromStore construct
func NewPromStore(metrics ...prometheus.Collector) *PromMonitor {
	prom := &PromMonitor{
		PullTaskCounterVec: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "uTask",
			Subsystem: "consumer",
			Name:      "pull_task_total",
			Help:      "pull task from task store",
		}, []string{"appID", "sid", "cid", "type"}),
		HandleTaskCounterVec: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "uTask",
			Subsystem: "consumer",
			Name:      "handel_task_total",
			Help:      "handel task total",
		}, []string{"appID", "sid", "cid", "type", "status"}),
		HandleTaskDuration: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "uTask",
			Subsystem: "consumer",
			Name:      "handel_task_duration",
			Help:      "handle task duration(ms)",
		}, []string{"appID", "sid", "cid", "type", "status"}),
		RetryTaskCounterVec: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "uTask",
			Subsystem: "consumer",
			Name:      "retry_task_total",
			Help:      "retry task handler total",
		}, []string{"appID", "sid", "cid", "type"}),
		RequestCounterVer: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "uTask",
			Subsystem: "producer",
			Name:      "request_task_total",
			Help:      "request task total",
		}, []string{"type"}),
		ResponseLatency: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "uTask",
			Subsystem: "producer",
			Name:      "response_latency",
			Help:      "response latency duration(ms)",
		}, []string{"type"}),
	}
	prometheus.MustRegister(prom.PullTaskCounterVec)
	prometheus.MustRegister(prom.HandleTaskCounterVec)
	prometheus.MustRegister(prom.HandleTaskDuration)
	prometheus.MustRegister(prom.RetryTaskCounterVec)
	prometheus.MustRegister(prom.RequestCounterVer)
	prometheus.MustRegister(prom.ResponseLatency)
	return prom
}

// PullTask pull task metrics
func (prom *PromMonitor) PullTask(cid string, task task.Tasker) {
	prom.PullTaskCounterVec.WithLabelValues(task.GetAppID(), task.GetSID(), cid, task.GetType()).Inc()
}

// HandleTask handle task metrics
func (prom *PromMonitor) HandleTask(cid string, task task.Tasker) {
	status := 1
	if err := task.GetLastError(); err != nil {
		status = 0
	}
	prom.HandleTaskCounterVec.WithLabelValues(task.GetAppID(), task.GetSID(), cid, task.GetType(), strconv.Itoa(status)).Inc()
	if task.GetLastExecTime() != 0 {

		prom.HandleTaskCounterVec.WithLabelValues(task.GetAppID(), task.GetSID(), cid, task.GetType(), strconv.Itoa(status)).Add(float64(task.GetLastExecTime()))
	}
}

// Retries consumer retries
func (prom *PromMonitor) Retries(cid string, task task.Tasker) {
	prom.RetryTaskCounterVec.WithLabelValues(task.GetAppID(), task.GetSID(), cid, task.GetType()).Inc()
}

// Request producer request
func (prom *PromMonitor) Request(processType string) {
	prom.RequestCounterVer.WithLabelValues(processType).Inc()
}

// Latency producer process latency
func (prom *PromMonitor) Latency(processType string, duration time.Duration) {
	prom.ResponseLatency.WithLabelValues(processType).Add(float64(duration))
}
