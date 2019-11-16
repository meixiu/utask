package coder

import "github.com/meixiu/utask/task"

// Coder 任务对象序列化和反序列化接口
type Coder interface {
	// Encode 将任务序列化为二进制
	Encode(task.Tasker) ([]byte, error)
	// Decode 将二进制数据反序列化
	Decode([]byte) (task.Tasker, error)
}
