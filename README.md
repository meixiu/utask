![logo](docs/image/1_06.png)

# UTask 

方便可靠的异步任务处理系统

利用`UTask`业务方将不需要实现一套`队列 + 脚本`的架构, 直接实现`http接口`, 即可完成异步任务的执行, 跟踪和监控

可以广泛应用于各种异步任务的场景, 如:

- 游戏平台支付成功后通知游戏发货
- 定时通知游戏开启或者关闭一个活动
- 电商平台中下单后15分钟未付款时关闭订单


## 功能概述

- 异步调用HTTP接口
- 延迟调用 
- 定时调用 
- 异常重试机制
- 日志查询和监控

## TODO
- `Tasker` gRpc任务类型
- `TaskStorer` Kafka实现
- `TaskStorer` Redis Cluster实现

## 快速开始
TODO

## 测试第三方服务

```
cd test-third
go run main.go
```
- GET接口: 127.0.0.1:8021/test/get 
- POST接口: 127.0.0.1:8021/test/post 