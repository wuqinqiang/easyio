// Copyright (c) 2023 Remember
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package easyio

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type Conn interface {
	net.Conn

	Fd() int
	Flush() error
}

var _ Conn = (*conn)(nil)

type conn struct {
	fd          int
	localAddr   net.Addr
	rAddr       net.Addr
	network     string
	writeBuffer []byte

	closed atomic.Bool
	mux    sync.RWMutex

	poller *Poller
}

func (c *conn) Flush() error {
	if c.closed.Load() {
		return net.ErrClosed
	}
	c.mux.Lock()
	if len(c.writeBuffer) == 0 {
		c.mux.Unlock()
		return nil
	}
	n, err := c.Write(c.writeBuffer)
	if err != nil && !errors.Is(err, syscall.EINTR) && !errors.Is(err, syscall.EAGAIN) {
		c.Close()
		c.poller.DeleteConn(c)
		return err
	}
	if n <= 0 {
		return nil
	}
	if n < len(c.writeBuffer) {
		// todo opt
		c.writeBuffer = c.writeBuffer[n:]
		c.mux.Unlock()
		return nil
	}
	// reset to read
	c.writeBuffer = nil
	c.resetRead()
	c.mux.Unlock()

	return nil
}

func (c *conn) resetRead() {
	if !c.closed.Load() {
		c.poller.ModRead(c.fd)
	}
}

func (c *conn) Read(b []byte) (n int, err error) {
	n, err = syscall.Read(c.fd, b)
	fmt.Println("Conn Read:", n)
	return
}
func (c *conn) Fd() int {
	return c.fd
}

func (c *conn) Write(b []byte) (n int, err error) {
	n, err = syscall.Write(c.fd, b)
	return
}

func (c *conn) Close() error {
	fmt.Println("closed:", c.Fd())
	c.closed.Store(true)
	return syscall.Close(c.fd)
}

func (c *conn) LocalAddr() net.Addr {
	return c.localAddr
}

func (c *conn) RemoteAddr() net.Addr {
	return c.rAddr
}

func (c *conn) SetDeadline(t time.Time) error {
	return nil
}

func (c *conn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *conn) SetWriteDeadline(t time.Time) error {
	return nil
}