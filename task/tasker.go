package task

import "context"

// Tasker 是任务接口，规范任务行为
type Tasker interface {
	//Init 初始化任务参数
	Init(sid string)
	//Validate 验证参数
	Validate() error
	//GetType 获取任务类型
	GetType() string
	//GetAppID 获取业务ID
	GetAppID() string
	//GetID 获取任务ID
	GetID() string
	//GetSID 获取任务来源Server ID
	GetSID() string
	//GetExpectTime 获取任务期望执行时间
	GetExpectTime() int64
	//Run 运行任务
	Run(ctx context.Context, token string) (result interface{}, err error)
	//SetProcessing 设置任务处于待处理队列中
	SetProcessing()
	//IsProcessing 获取任务是否在待处理队列中
	IsProcessing() bool
	//IncreaseTimes 增加出错次数
	IncreaseTimes()
	//GetTimes 获取执行次数
	GetTimes() int64
	//GetNextTime 获取下次执行时间
	GetNextTime() int64
	//MaxRetryTimes 获取最大执行次数
	MaxRetryTimes() int
	//Timeout 获取超时时间
	Timeout() int64
	//GetContent 获取任务内容
	GetContent() string
	//GetLastResult 获取最后一次结果
	GetLastResult() string
	//GetLastError 获取最后一次错误
	GetLastError() error
	//GetExecTime 获取任最后一次务执行花费时间
	GetLastExecTime() int64
}

// Register 注册任务表类型
type Register map[string]func() Tasker

// defaultRegister 注册任务表
var defaultRegister = Register{}

// Reg 根据任务类型名注册一个任务类型
func Reg(t string, task func() Tasker) {
	defaultRegister[t] = task
}

// Lookup 根据任务类型名查询一个任务类型
func Lookup(t string) Tasker {
	m, ok := defaultRegister[t]
	if !ok || m == nil {
		return nil
	}
	return m()
}

// GetRegister 获取完整的注册任务表
func GetRegister() Register {
	return defaultRegister
}
