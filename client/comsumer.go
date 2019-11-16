package client

import (
	"context"
)

// Consumer 一个消费者接口
type Consumer interface {
	// Start 开始消费者任务
	Start() error
	// Stop 结束消费者任务
	Stop(ctx context.Context) error
}
