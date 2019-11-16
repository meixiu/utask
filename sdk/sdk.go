package sdk

// Pusher 任务推送接口
type Pusher interface {
	// Register 注册一个业务ID和密钥
	Register(appId string, appSecret string)
	// Push 推送一个任务
	Push(task interface{}) (taskId string, err error)
}

// Checker 任务认证接口
type Checker interface {
	// Check 校验任务Token
	Check(taskId string, token string) error
}
