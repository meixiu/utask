package task

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/meixiu/utask/log"

	"github.com/meixiu/httpclient"

	"github.com/google/uuid"
)

func init() {
	Reg("http", func() Tasker {
		return &HttpTask{}
	})
}

var (
	headerUTaskId    = "U-Task-Id"
	headerUTaskToken = "U-Task-Token"
)

// Resp 接口返回包
type HttpResp struct {
	Code    int    `json:"code"`    // 状态码
	Message string `json:"message"` // 提示信息
}

// Http任务
type HttpTask struct {
	SID          string        `json:"sid"`         // Server ID
	ID           string        `json:"id"`          // Task Id
	CreateTime   int64         `json:"create_time"` // 创建时间
	NextTime     int64         `json:"next_time"`   // 下次执行时间
	Times        int64         `json:"times"`       // 执行次数
	Processing   int           `json:"processing"`  // 执行中
	lastResult   string        // 最后一次执行结果
	lastError    error         // 是后一次执行错误
	lastExecTime time.Duration // 最后一次执行花费时间

	ExpectTime  int64  `json:"expect_time"`  // 等于0:立即执行; 小于一年:延时执行; 其他值:定时执行
	AppID       string `json:"app_id"`       // 业务ID
	URL         string `json:"url"`          // 请求地址
	Method      string `json:"method"`       // GET|POST
	ContentType string `json:"content_type"` // 默认为JSON
	Body        string `json:"body"`         // 请求原数据
}

func (t *HttpTask) Init(sid string) {
	t.ID = uuid.New().String()
	t.SID = sid
	t.CreateTime = time.Now().Unix()
	t.NextTime = 0
	t.Times = 0
}

func (t *HttpTask) Validate() error {
	if t.AppID == "" {
		return fmt.Errorf("incorrect parameter: %s", "app_id")
	}
	if t.URL == "" {
		return fmt.Errorf("incorrect parameter: %s", "url")
	}
	return nil
}

func (t HttpTask) GetType() string {
	return "http"
}

func (t *HttpTask) SetProcessing() {
	t.Processing = 1
}

func (t HttpTask) IsProcessing() bool {
	return t.Processing == 1
}

// GetExpectTime 返回预期执行时间
// ExpectTime等于0 立即执行;
// ExpectTime小于一年 延时执行;
// ExpectTime其他值 定时执行
func (t HttpTask) GetExpectTime() int64 {
	if t.ExpectTime < 3600*24*365 {
		return t.CreateTime + t.ExpectTime
	}
	return t.ExpectTime
}

func (t HttpTask) GetAppID() string {
	return t.AppID
}

func (t HttpTask) GetID() string {
	return t.ID
}

func (t HttpTask) GetSID() string {
	return t.SID
}

func (t *HttpTask) Run(ctx context.Context, token string) (result interface{}, err error) {
	log.Info("Run Task: ", t.ID, "SID: ", t.SID, "Data: ", *t)
	// 处理http请求
	client := httpclient.New()
	client.SetHeader(headerUTaskId, t.ID)
	client.SetHeader(headerUTaskToken, token)
	startTime := time.Now()
	var resp *httpclient.Response
	switch strings.ToUpper(t.Method) {
	case http.MethodPost:
		resp, err = client.PostJson(t.URL, t.Body)
	default:
		resp, err = client.Get(t.URL, nil)
	}
	t.lastExecTime = time.Now().Sub(startTime)
	if err != nil {
		t.lastError = err
		return nil, t.lastError
	}

	// 检测接口约定返回值
	t.lastResult = resp.String()
	res := &HttpResp{}
	if err := resp.Decode(res); err != nil {
		t.lastError = err
		return nil, t.lastError
	}

	// 错误码不等于0时表示失败
	if res.Code != 0 {
		err := fmt.Errorf("code=%d error=%s", res.Code, res.Message)
		t.lastError = err
		return nil, t.lastError
	}
	return res, nil
}

func (t HttpTask) GetTimes() int64 {
	return t.Times
}

func (t *HttpTask) IncreaseTimes() {
	t.Times += 1
	t.NextTime = time.Now().Unix() + t.Times*t.Times*1
}

func (t HttpTask) GetNextTime() int64 {
	return t.NextTime
}

func (t HttpTask) MaxRetryTimes() int {
	return 6
}

func (t HttpTask) Timeout() int64 {
	return 300
}

func (t HttpTask) GetContent() string {
	content := map[string]interface{}{
		"url":          t.URL,
		"method":       t.Method,
		"body":         t.Body,
		"content_type": t.ContentType,
		"expect_time":  t.ExpectTime,
	}
	b, _ := json.Marshal(content)
	return string(b)
}

func (t HttpTask) GetLastResult() string {
	return t.lastResult
}

func (t HttpTask) GetLastError() error {
	return t.lastError
}

func (t HttpTask) GetLastExecTime() int64 {
	return t.lastExecTime.Milliseconds()
}
