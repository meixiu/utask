package client

import (
	"context"
	"time"
	"utask/app"
	"utask/log"
	"utask/monitor"
	"utask/store"
	"utask/task"
)

const (
	//默认拉取长度
	FetchProcessStoreSize = 10
)

// ChanClient 利用chan的消费者类型
type ChanClient struct {
	id       string        // 队列唯一标志
	interval time.Duration // 处理时间间隔
	//timeout  time.Duration // 处理超时时间 默认：300个处理时间间隔

	taskStore    store.TaskStorer        // 获取任务数据源
	secretStore  store.SecretStorer      // 安全校验数据源
	processStore store.ProcessStorer     // 任务处理数据源
	logStore     store.LogStorer         // 任务日志数据源
	monitor      monitor.ConsumerMonitor // 任务监控

	stop    chan struct{}    // 处理停止信号
	suspend chan bool        // 处理暂停信号
	process chan struct{}    // 当前处理计数
	waits   chan task.Tasker // 等待处理队列

	maxWaits   int // 队列等待数 默认：128
	maxProcess int // 处理并发数 默认：64
}

// NewChanClient 返回一个消费客户端
func NewChanClient(id string, opts Options) Consumer {
	return &ChanClient{
		id:       id,
		interval: time.Second,

		taskStore:    opts.TaskStore,
		processStore: opts.ProcessStore,
		logStore:     opts.LogStore,
		secretStore:  opts.SecretStore,
		monitor:      opts.Monitor,

		stop:       make(chan struct{}),
		suspend:    make(chan bool, 1),
		waits:      make(chan task.Tasker, app.Config.Cli.MaxWaits),
		process:    make(chan struct{}, app.Config.Cli.MaxProcess),
		maxWaits:   app.Config.Cli.MaxWaits,
		maxProcess: app.Config.Cli.MaxProcess,
	}
}

// Start 开始队列, 只能开始一次
func (c *ChanClient) Start() error {
	//暂停信号获取
	async(func() {
		t := time.NewTimer(0)
		defer t.Stop()
		for {
			<-t.C
			t.Reset(c.interval)

			s, err := c.IsSuspend()
			if err != nil {
				log.Error("client suspend err: ", err)
				continue
			}
			c.suspend <- s
		}
	}, nil)

	suspend := <-c.suspend //同步当前状态

	abnormalTimer := time.NewTimer(0) //出错重试、超时重试、延时任务 定时获取
	defer abnormalTimer.Stop()

	normalTimer := time.NewTimer(0) //获取任务数据源任务 计时器
	defer normalTimer.Stop()

	for {
		//操作信号
		select {
		case suspend = <-c.suspend: //暂停信号
		case <-c.stop: //停止信号 接收成功，c.waits不会再处理, normal、abnormal增加停止
			log.Info("client stopped")
			return nil
		default: //当前没有信号，快速运行业务逻辑
		}
		//检测暂停状态
		if suspend {
			log.Info("client suspend for ", c.interval)
			<-time.After(c.interval)
			continue
		}
		//业务逻辑
		//这里需要异步获取，如果获取任务是同步的，插入可能阻塞
		select {
		case <-normalTimer.C: //获取任务数据源任务 c.waits <- normal
			async(func() {
				nextTime := c.interval
				defer func() {
					normalTimer.Reset(nextTime)
				}()
				log.Info("client normal info")

				count, err := c.Normal()
				if err != nil {
					log.Error("client normal err: ", err)
					return
				}

				if count > 0 {
					nextTime = 0
				}
			}, nil)
		case <-abnormalTimer.C: //出错重试、超时重试、延时任务 c.waits <- abnormal
			async(func() {
				defer func() {
					abnormalTimer.Reset(c.interval)
				}()
				log.Info("client abnormal info")

				_, err := c.Abnormal()
				if err != nil {
					log.Error("client abnormal error: ", err)
					return
				}
			}, nil)
		case item := <-c.waits: //处理队列 c.waits -> process
			async(func() { _ = c.Dispose(item) }, c.process)
		}
	}
}

// Stop 停止队列, 只能停止一次
func (c *ChanClient) Stop(ctx context.Context) error {
	c.stop <- struct{}{}                               //发送成功才会继续
	tempProcess := make(chan struct{}, len(c.waits)+1) //尽可能全部重置

	t := time.NewTicker(c.interval / 10)
	defer t.Stop()
	for {
		if len(c.waits)+len(c.process)+len(tempProcess) == 0 {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case item := <-c.waits:
			async(func() { _ = c.Reset(item) }, tempProcess)
		case <-t.C: //尽快重新进入循环，防止超时
		}
	}
}

// Add 接收队列项到内存channel
func (c *ChanClient) Add(items ...task.Tasker) {
	//此处可能有性能问题, 因此丢弃掉超时的
	//结果是这些数据会在超时任务区处理
	t := time.NewTimer(0)
	defer t.Stop()
	<-t.C
	for _, item := range items {
		item := item //循环变量->局部变量
		t.Reset(time.Duration(item.Timeout()) * time.Second)
		select {
		case c.waits <- item:
		case <-t.C:
		}
	}
}

// Insert 接收参数并存入数据处理区
func (c *ChanClient) Normal() (count int, err error) {
	item, err := c.taskStore.LPop()
	if err != nil {
		return 0, err
	}
	if item == nil {
		return 0, nil
	}
	//是否立即处理
	if int64(c.interval/time.Second) > item.GetExpectTime()-time.Now().Unix() {
		item.SetProcessing()
	}
	//插入处理数据源
	err = c.processStore.Insert(c.id, item)
	if err != nil {
		ok, errPush := c.taskStore.RPush(item)
		if errPush != nil || !ok {
			//这条记录需要重发
			log.Error("client normal push err, task: ", item, "insert err: ", err, " push back status: ", ok, " push back err: ", errPush)
		}
		return 0, err
	}
	//二次确认
	if cs, ok := c.processStore.(store.ConfirmTaskStorer); ok {
		errConfirm := cs.Confirm(item.GetID())
		if errConfirm != nil {
			//这条记录会重发
			log.Error("client normal confirm err, task: ", item, "insert err: ", err, " push back status: ", ok, " push back err: ", errConfirm)
		}
	}
	if item.IsProcessing() {
		c.Add(item)
	}
	return 1, nil
}

// Abnormal 获取任务处理区数据 (出错重试、超时重试、延时任务)
func (c *ChanClient) Abnormal() (count int, err error) {
	items, err := c.processStore.Get(c.id, FetchProcessStoreSize)
	if err != nil {
		return 0, err
	}
	c.Add(items...)
	return len(items), nil
}

// Reset 重置下次处理时间
func (c *ChanClient) Reset(item task.Tasker) error {
	//设置下次处理时间
	item.IncreaseTimes()
	_, err := c.processStore.Update(c.id, item)
	if err != nil {
		return err
	}
	return nil
}

// Dispose 调用任务处理
func (c *ChanClient) Dispose(item task.Tasker) (err error) {
	tid := item.GetID()
	log.Info("client dispose item: ", tid, item)

	timeout := time.Duration(item.Timeout()) * time.Second
	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()

	token, err := c.secretStore.Generate(tid, timeout)
	if err != nil {
		return err
	}
	_, err = item.Run(ctx, token)
	_ = c.Log(item)
	if err != nil {
		log.Error("client dispose run err: ", err, item)
		resetErr := c.Reset(item)
		log.Error("client dispose reset err: ", resetErr, item)
		return err
	}
	_, err = c.Delete(item)
	log.Info("client dispose del: ", tid, item, err)
	return nil
}

// Delete 从任务处理区删除
func (c *ChanClient) Delete(item task.Tasker) (bool, error) {
	return c.processStore.Delete(c.id, item.GetID())
}

// Log 插入日志总表
func (c *ChanClient) Log(item task.Tasker) error {
	return c.logStore.Log(c.id, item)
}

// IsSuspend 判断任务处理是否暂停
func (c *ChanClient) IsSuspend() (bool, error) {
	return false, nil
}

// async 简单的控制协程数量的函数
func async(f func(), c chan struct{}) {
	if c == nil || cap(c) == 0 {
		go f()
		return
	}
	c <- struct{}{}
	go func() {
		defer func() { <-c }()
		f()
	}()
	return
}
