package server

import (
	"context"
	"net/http"
	"time"

	"github.com/meixiu/utask/app"
	"github.com/meixiu/utask/monitor"
	"github.com/meixiu/utask/store"
	"github.com/meixiu/utask/task"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gin-gonic/gin"
)

const (
	errCodeDataType   = 1001 // 任务type不正确
	errCodeDataBind   = 1002 // 参数绑定错误
	errCodeParams     = 1003 // 参数错误
	errCodePushQueue  = 1004 // 入队列错误
	errCodeCheckToken = 2001 // token校验错误
)

// NewHttpServer http server cli
func NewHttpServer(id string, opts Options) Producer {
	return &HttpServer{
		ID:          id,
		Addr:        app.Config.Server.Addr,
		TaskStore:   opts.TaskStore,
		SecretStore: opts.SecretStore,
		Monitor:     opts.Monitor,
	}
}

// HttpResp resp
type HttpResp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// HttpServer server
type HttpServer struct {
	ID          string
	TaskStore   store.TaskStorer
	SecretStore store.SecretStorer
	Monitor     monitor.ProducerMonitor

	Server *http.Server
	Addr   string
}

// MwPrometheusHttp middleware
func (s *HttpServer) MwPrometheusHttp(c *gin.Context) {
	start := time.Now()
	s.Monitor.Request("http")
	c.Next()
	// after request
	end := time.Now()
	d := end.Sub(start) / time.Millisecond
	s.Monitor.Latency("http", d)
}

// ListenAndServe listens
func (s *HttpServer) ListenAndServe() error {
	router := gin.Default()

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	api := router.Group("api").Use(s.MwPrometheusHttp)

	api.POST("/task/:type", s.Handle)
	api.POST("/check", s.Check)

	s.Server = &http.Server{
		Addr:           s.Addr,
		Handler:        router,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	return s.Server.ListenAndServe()
}

// Shutdown Shutdown
func (s *HttpServer) Shutdown(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}

// Handle handle
func (s *HttpServer) Handle(ctx *gin.Context) {
	t := ctx.Param("type")
	tasker := task.Lookup(t)
	if tasker == nil {
		ctx.JSON(http.StatusOK, HttpResp{Code: errCodeDataType, Message: "data type error"})
		return
	}
	if err := ctx.Bind(tasker); err != nil {
		ctx.JSON(http.StatusOK, HttpResp{Code: errCodeDataBind, Message: "data bind error"})
		return
	}
	// 校验参数
	if err := tasker.Validate(); err != nil {
		ctx.JSON(http.StatusOK, HttpResp{Code: errCodeParams, Message: err.Error()})
		return
	}
	ok, err := Push(s.ID, s.TaskStore, tasker)
	if err != nil || !ok {
		ctx.JSON(http.StatusOK, HttpResp{Code: errCodePushQueue, Message: "add queue error"})
		return
	}
	ctx.JSON(http.StatusOK, HttpResp{
		Code:    0,
		Message: "success",
		Data: gin.H{
			"task_id": tasker.GetID(),
		},
	})
	return
}

// Check token checks
func (s *HttpServer) Check(ctx *gin.Context) {
	dataCheck := &DataCheck{}
	err := ctx.ShouldBind(dataCheck)
	if err != nil {
		ctx.JSON(http.StatusOK, HttpResp{Code: errCodeDataBind, Message: "data bind error"})
		return
	}
	ok, err := s.SecretStore.Check(dataCheck.TaskID, dataCheck.Token)
	if err != nil || !ok {
		ctx.JSON(http.StatusOK, HttpResp{Code: errCodeCheckToken, Message: "check token error"})
		return
	}
	ctx.JSON(http.StatusOK, HttpResp{Code: 0, Message: "success"})
	return
}

// DataCheck token check struct
type DataCheck struct {
	TaskID string `json:"task_id" form:"task_id"`
	Token  string `json:"token" form:"token"`
}
