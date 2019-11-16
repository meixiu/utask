package app

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/meixiu/utask/pkg/network"

	"gopkg.in/yaml.v2"
)

type config struct {
	Debug   bool   `json:"debug" yaml:"debug"`
	Version string `json:"version" yaml:"version"`
	Server  struct {
		Addr string `json:"addr" yaml:"addr"`
		Url  string `json:"url"`
	}
	Db struct {
		Driver        string `json:"driver" yaml:"driver"`
		Source        string `json:"source" yaml:"source"`
		MaxOpenConns  int    `json:"max_open_conns" yaml:"max_open_conns"`
		MaxIdleConns  int    `json:"max_idle_conns" yaml:"max_idle_conns"`
		MaxRetryTimes int    `json:"max_retry_times" yaml:"max_retry_times"`
		MaxLockTime   int64  `json:"max_lock_time" yaml:"max_lock_time"`
	}
	Redis struct {
		Addr     string `json:"addr" yaml:"addr"`
		Password string `json:"password" yaml:"password"`
		Db       int    `json:"db" yaml:"db"`
	}
	Cli struct {
		MaxWaits   int `json:"max_waits" yaml:"max_waits"`
		MaxProcess int `json:"max_process" yaml:"max_process"`
	}
}

var (
	Config *config
)

func init() {
	loadConfig()
}

func loadConfig() {
	file := flag.String("c", "config/dev.yaml", "config file path")
	flag.Parse()
	data, err := ioutil.ReadFile(*file)
	if err != nil {
		log.Fatal(err)
	}
	if err := yaml.Unmarshal(data, &Config); err != nil {
		log.Fatal(err)
	}
}

// ServerId 生产者进程ID
func ServerId() string {
	return fmt.Sprintf("%s%s", network.InternalIP(), Config.Server.Addr)
}

// ClientId 消费者进程ID
func ClientId() string {
	return fmt.Sprintf("%s%s", network.InternalIP(), Config.Server.Addr)
}
