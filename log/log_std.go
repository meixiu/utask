package log

import (
	"fmt"
	"strings"
	"time"
)

// StdLog 标准输出日志输出类
type StdLog struct {
}

// Debug 打印Debug信息
func (l StdLog) Debug(info ...interface{}) {
	l.Println("Debug", info...)
}

// Info 打印Info信息
func (l StdLog) Info(info ...interface{}) {
	l.Println("Info", info...)
}

// Warning 打印Warning信息
func (l StdLog) Warning(info ...interface{}) {
	l.Println("Warning", info...)
}

// Error 打印Error信息
func (l StdLog) Error(info ...interface{}) {
	l.Println("Error", info...)
}

// Println 输出附带时间日志信息
func (l StdLog) Println(level string, data ...interface{}) {
	m := strings.Builder{}
	m.Grow(1024)
	m.WriteString(time.Now().Format("2006-01-02 15:04:05.000\t"))
	m.WriteString(level)
	m.WriteRune('\t')
	for _, v := range data {
		m.WriteString(fmt.Sprintf("%+v\t", v))
	}
	fmt.Println(m.String())
}
