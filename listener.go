// Copyright (c) 2023 Remember
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build linux

package easyio

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

var _ net.Listener = (*Listener)(nil)

type Listener struct {
	once   sync.Once
	ln     net.Listener
	addr   net.Addr //local addr
	engine *Engine
}

func (ln *Listener) Close() error {
	ln.once.Do(func() {
		if ln.ln != nil {
			ln.ln.Close()
		}
	})
	return nil
}

func (ln *Listener) Addr() net.Addr {
	return ln.addr
}

func (ln *Listener) Accept() (net.Conn, error) {
	conn, err := ln.ln.Accept()
	if err != nil {
		var ne net.Error
		if ok := errors.As(err, &ne); ok && ne.Timeout() {
			fmt.Println("Accept failed", err)
			time.Sleep(time.Second)
			return nil, nil
		}
		return nil, err
	}

	if conn == nil {
		return nil, nil
	}
	c, ok := conn.(Conn)
	if !ok {
		var err error
		c, err = dupStdConn(conn)
		if err != nil {
			conn.Close()
			return nil, err
		}
	}
	return c, nil
}
