// Copyright (c) 2023 Remember
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package easyio

import (
	"context"
	"fmt"
	"syscall"
)

const (
	ErrEvents = syscall.EPOLLERR | syscall.EPOLLHUP | syscall.EPOLLRDHUP

	ReadEvents = syscall.EPOLLIN | syscall.EPOLLPRI

	WriteEvents = syscall.EPOLLOUT
)

type EventHandler interface {
	OnOpen(c Conn) context.Context
	OnRead(ctx context.Context, c Conn)
	OnClose(ctx context.Context, c Conn)
}

var _ EventHandler = (*eventHandler)(nil)

type eventHandler struct{}

func (d *eventHandler) OnOpen(c Conn) context.Context {
	return context.Background()
}

func (d *eventHandler) OnClose(_ context.Context, c Conn) {
	fmt.Printf("[OnClose] conn: %d closed\n", c.Fd())
}

func (d *eventHandler) OnRead(ctx context.Context, c Conn) {
	// todo set reader buffer
	b := make([]byte, 1024)
	n, err := c.Read(b[:cap(b)])
	if err != nil {
		fmt.Println("OnRead err:", err)
	}
	fmt.Println("read data: ", string(b[:n]))
	n, err = c.Write(b[:n])
	fmt.Println("write len:", n)
}
