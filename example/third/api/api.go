package api

import (
	"net/http"
	"time"

	"github.com/meixiu/utask/log"
	"github.com/meixiu/utask/sdk"

	"github.com/gin-gonic/gin"
)

var (
	// checker
	checker sdk.Checker = sdk.NewHttpCheck("http://127.0.0.1:8020")
)

func TestGet(c *gin.Context) {
	taskId := c.GetHeader("U-Task-Id")
	token := c.GetHeader("U-Task-Token")
	if err := checker.Check(taskId, token); err != nil {
		retError(c, 1001, err)
		return
	}
	retData(c, gin.H{
		"method": "get",
		"time":   time.Now(),
	})
}

func TestPost(c *gin.Context) {
	taskId := c.GetHeader("U-Task-Id")
	token := c.GetHeader("U-Task-Token")
	if err := checker.Check(taskId, token); err != nil {
		retError(c, 1001, err)
		return
	}
	retData(c, gin.H{
		"method": "post",
		"time":   time.Now(),
	})
}

// recordLog 记录请求日志
func recordLog(c *gin.Context, resp *sdk.HttpCheckResp) {
	log.Info(c.Request.URL, resp)
}

// retError 返回错误
func retError(c *gin.Context, code int, err error) {
	resp := &sdk.HttpCheckResp{
		Code:    code,
		Message: err.Error(),
		Data:    nil,
	}
	recordLog(c, resp)
	c.JSON(http.StatusOK, resp)
}

// retData 返回数据
func retData(c *gin.Context, data interface{}) {
	resp := &sdk.HttpCheckResp{
		Message: "success",
		Data:    data,
	}
	recordLog(c, resp)
	c.JSON(http.StatusOK, resp)
}
