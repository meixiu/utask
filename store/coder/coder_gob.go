package coder

import (
	"bytes"
	"encoding/gob"
	"utask/task"
)

// GobCoder 是实现对任务进行Gob序列化、反序列化的
type GobCoder struct {
}

// NewGobCoder 返回一个GobCoder对象
func NewGobCoder() *GobCoder {
	return &GobCoder{}
}

// Encode 使用Gob序列化一个任务
func (c GobCoder) Encode(task task.Tasker) ([]byte, error) {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(&task)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decode 使用Gob反序列化出一个任务
func (c GobCoder) Decode(data []byte) (task task.Tasker, err error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&task)
	if err != nil {
		return nil, err
	}
	return task, nil
}
