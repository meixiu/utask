package store

import (
	"encoding/gob"
	"time"

	"github.com/meixiu/utask/store/coder"
	"github.com/meixiu/utask/task"
)

// TaskStorer 任务数据源
type TaskStorer interface {
	//LPop 从数据源获取一个任务
	LPop() (task.Tasker, error)
	//RPush 增加一个任务到数据源
	RPush(task task.Tasker) (bool, error)
}

// ConfirmTaskStorer 二次确认数据源
type ConfirmTaskStorer interface {
	TaskStorer
	//Confirm 二次确认
	Confirm(tid string) error
}

// TaskStorer 任务处理区
type ProcessStorer interface {
	//Get 根据客户端、拉取个数从任务处理区拉取任务
	Get(cid string, size int) ([]task.Tasker, error)
	//Insert 插入一条任务到处理区
	Insert(cid string, task task.Tasker) (err error)
	//Update 使用cid和任务进行更新
	Update(cid string, task task.Tasker) (bool, error)
	//Delete 根据cid和任务id进行删除
	Delete(cid string, id string) (bool, error)
}

// StealProcessStorer 能窃取任务的数据源
type StealProcessStorer interface {
	ProcessStorer
	// Mark 标记一个任务可被窃取
	Mark(cid string, task task.Tasker) (bool, error)
	// Steal 窃取size个任务
	Steal(cid string, size int) (int64, error)
}

// LogStorer 任务日志区
type LogStorer interface {
	//Log 插入一条任务日志到日志区
	Log(cid string, task task.Tasker) error
}

// SecretStorer 任务日志区
type SecretStorer interface {
	//Generate 生成一个token
	Generate(tid string, lifetime time.Duration) (token string, err error)
	//Check 校验并删除一个token
	Check(tid, token string) (ok bool, err error)
}

// defaultCoder 默认的编码器
var defaultCoder coder.Coder

// SetCoder 设置任务序列化反序列化方法
func SetCoder(c coder.Coder) {
	defaultCoder = c
}

// Encode 调用defaultCoder序列化任务
func Encode(task task.Tasker) (data []byte, err error) {
	return defaultCoder.Encode(task)
}

// Decode 调用defaultCoder反序列化任务
func Decode(data []byte) (task task.Tasker, err error) {
	return defaultCoder.Decode(data)
}

func init() {
	//任务对象注册gob
	for _, v := range task.GetRegister() {
		gob.Register(v())
	}
	SetCoder(coder.NewGobCoder())
}
