// Copyright 2019 Yoozoo Authors. All Rights Reserved.
// @Description: options

package server

import (
	"utask/monitor"
	"utask/store"
)

// Options option info
type Options struct {
	TaskStore   store.TaskStorer
	SecretStore store.SecretStorer
	Monitor     monitor.ProducerMonitor
}

// Option Option
type Option func(*Options)

// NewOptions construct
func NewOptions(opts ...Option) Options {
	opt := Options{
		TaskStore:   store.DefaultRedisStore,
		SecretStore: store.DefaultRedisStore,
		Monitor:     monitor.DefaultPromMonitor,
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

// Monitor monitor
func Monitor(m monitor.ProducerMonitor) Option {
	return func(o *Options) {
		o.Monitor = m
	}
}
