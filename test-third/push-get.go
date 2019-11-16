package main

import (
	"fmt"
	"utask/sdk"
)

func main() {
	pusher := sdk.NewHttpPush("http://127.0.0.1:8020/")
	taskId, err := pusher.Push(sdk.HttpPushReq{
		AppID: "100",
		URL:   "http://127.0.0.1:8021/test/get",
	})
	fmt.Println(taskId, err)
}
