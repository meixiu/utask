package main

import (
	"encoding/json"
	"fmt"

	"github.com/meixiu/utask/sdk"
)

func main() {
	pusher := sdk.NewHttpPush("http://127.0.0.1:8020/")
	data := map[string]string{
		"name1": "value1",
		"name2": "value2",
	}
	body, _ := json.Marshal(data)
	taskId, err := pusher.Push(sdk.HttpPushReq{
		AppID:  "100",
		URL:    "http://127.0.0.1:8021/test/post",
		Method: "POST",
		Body:   string(body),
	})
	fmt.Println(taskId, err)
}
