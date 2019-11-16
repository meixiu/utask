package sdk

import (
	"fmt"
	"strings"

	"github.com/meixiu/httpclient"
)

const (
	PushPath  = "/api/task/http" // 任务推送接口
	CheckPath = "/api/check"     // 任务认证接口
)

type (
	// HttpPushReq HTTP任务请求参数
	HttpPushReq struct {
		ExpectTime  int64  `json:"expect_time"`  // 等于0:立即执行; 小于一年:延时执行; 其他值:定时执行
		AppID       string `json:"app_id"`       // 业务ID
		URL         string `json:"url"`          // 请求地址
		Method      string `json:"method"`       // GET|POST
		ContentType string `json:"content_type"` // 默认为JSON
		Body        string `json:"body"`         // 请求原数据
	}

	// HttpCheckReq HTTP任务认证参数
	HttpCheckReq struct {
		TaskId string `json:"task_id"` // 任务ID
		Token  string `json:"token"`   // 任务TOKEN
	}

	// HttpPushResp HTTP任务返回参数
	HttpPushResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			TaskId string `json:"task_id"`
		} `json:"data"`
	}

	// HttpCheckResp HTTP认证返回参数
	HttpCheckResp struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}
)

// NewHttpPush
func NewHttpPush(url string) *HttpPush {
	url = strings.TrimSuffix(url, "/")
	return &HttpPush{Url: url}
}

// HttpPush 是HTTP任务推送接口实现
type HttpPush struct {
	Url       string // 服务Url
	appId     string // 业务ID
	appSecret string // 业务密钥
}

func (h *HttpPush) Register(appId string, appSecret string) {
	h.appId = appId
	h.appSecret = appSecret
}

func (h *HttpPush) Push(task interface{}) (string, error) {
	uri := h.Url + PushPath
	client := httpclient.New()
	resp, err := client.PostJson(uri, task)
	if err != nil {
		return "", err
	}
	data := &HttpPushResp{}
	if err := resp.Decode(data); err != nil {
		return "", err
	}
	if data.Code != 0 {
		return "", fmt.Errorf("code=%d, message=%s", data.Code, data.Message)
	}
	return data.Data.TaskId, nil
}

func NewHttpCheck(url string) *HttpCheck {
	url = strings.TrimSuffix(url, "/")
	return &HttpCheck{Url: url}
}

// HttpCheck 是HTTP任务认证接口实现
type HttpCheck struct {
	Url string // 服务Url
}

func (h *HttpCheck) Check(taskId string, token string) error {
	uri := h.Url + CheckPath
	client := httpclient.New()
	resp, err := client.PostJson(uri, HttpCheckReq{
		TaskId: taskId,
		Token:  token,
	})
	if err != nil {
		return err
	}
	data := &HttpCheckResp{}
	if err := resp.Decode(data); err != nil {
		return err
	}
	if data.Code != 0 {
		return fmt.Errorf("code=%d, message=%s", data.Code, data.Message)
	}
	return nil
}
