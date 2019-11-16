// Copyright 2019 Yoozoo Authors. All Rights Reserved.
// @Description: options

package client

import (
	"utask/monitor"
	"utask/store"
)

// Options option info
type Options struct {
	TaskStore    store.TaskStorer
	SecretStore  store.SecretStorer
	ProcessStore store.ProcessStorer
	LogStore     store.LogStorer
	Monitor      monitor.ConsumerMonitor
}

// Option Option
type Option func(*Options)

// NewOptions construct
func NewOptions(opts ...Option) Options {
	opt := Options{
		TaskStore:    store.DefaultRedisStore,
		SecretStore:  store.DefaultRedisStore,
		ProcessStore: store.DefaultMysqlStore,
		LogStore:     store.DefaultMysqlStore,
		Monitor:      monitor.DefaultPromMonitor,
	}
	for _, o := range opts {
		o(&opt)
	}
	return opt
}

// TaskStore task store
func TaskStore(t store.TaskStorer) Option {
	return func(o *Options) {
		o.TaskStore = t
	}
}

// SecretStore secret store
func SecretStore(s store.SecretStorer) Option {
	return func(o *Options) {
		o.SecretStore = s
	}
}

// ProcessStore process store
func ProcessStore(p store.ProcessStorer) Option {
	return func(o *Options) {
		o.ProcessStore = p
	}
}

// LogStore log store
func LogStore(l store.LogStorer) Option {
	return func(o *Options) {
		o.LogStore = l
	}
}

// Monitor monitor
func Monitor(m monitor.ConsumerMonitor) Option {
	return func(o *Options) {
		o.Monitor = m
	}
}
