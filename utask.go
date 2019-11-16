package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"utask/app"
	"utask/client"
	"utask/server"
)

func main() {
	srvOpts := server.NewOptions()
	s := server.NewHttpServer(app.ServerId(), srvOpts)

	cliOpts := client.NewOptions()
	c := client.NewChanClient(app.ClientId(), cliOpts)

	go func() {
		_ = s.ListenAndServe()
	}()

	go func() {
		_ = c.Start()
	}()
	log.Println("Start @", app.Config.Version)
	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	log.Println("Shutdown ...", <-quit)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.Shutdown(ctx); err != nil {
			log.Println("Server Shutdown:", err)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := c.Stop(ctx); err != nil {
			log.Println("Queue Shutdown:", err)
		}
	}()
	wg.Wait()
	log.Println("Exiting")
}
