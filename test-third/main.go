package main

import (
	"net/http"
	"time"
	"utask/test-third/api"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.GET("/test/get", api.TestGet)
	router.POST("/test/post", api.TestPost)
	server := &http.Server{
		Addr:           ":8021",
		Handler:        router,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	server.ListenAndServe()
}
