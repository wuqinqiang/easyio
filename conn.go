// Copyright (c) 2023 Remember
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package easyio

import (
	"context"
	"errors"
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
	Context() context.Context
}

var _ Conn = (*conn)(nil)

type conn struct {
	ctx         context.Context
	fd          int
	localAddr   net.Addr
	rAddr       net.Addr
	network     string
	writeBuffer []byte

	closed atomic.Bool
	mux    sync.RWMutex

	poller *Poller
}

func (c *conn) Context() context.Context {
	return c.ctx
}
func (c *conn) write(b []byte) (int, error) {
	return syscall.Write(c.fd, b)
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

	n, err := c.write(c.writeBuffer)
	if err != nil && !errors.Is(err, syscall.EINTR) && !errors.Is(err, syscall.EAGAIN) {
		_ = c.Close()
		c.mux.Unlock()
		c.poller.removeConn(c)
		return err
	}
	if n <= 0 {
		c.mux.Unlock()
		return nil
	}

	byteBuffer := c.poller.e.GetByteBuffer()

	old := c.writeBuffer
	// handle remaining data
	if n < len(c.writeBuffer) {
		c.writeBuffer = byteBuffer.Get(len(c.writeBuffer) - n)
		copy(c.writeBuffer, old[n:])
		//give back []byte
		byteBuffer.Put(old)
		c.mux.Unlock()
		return nil
	}
	byteBuffer.Put(old)
	c.writeBuffer = nil
	c.mux.Unlock()

	// reset to read
	c.resetRead()
	return nil
}

func (c *conn) resetRead() {
	if !c.closed.Load() {
		c.poller.ModRead(c.fd)
	}
}

func (c *conn) Read(b []byte) (n int, err error) {
	n, err = syscall.Read(c.fd, b)
	return
}
func (c *conn) Fd() int {
	return c.fd
}

func (c *conn) Write(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	byteBuffer := c.poller.e.GetByteBuffer()

	c.mux.Lock()

	bufferLen := len(c.writeBuffer)
	if bufferLen == 0 {
		n, err := c.write(b)
		if err != nil && !errors.Is(err, syscall.EINTR) && !errors.Is(err, syscall.EAGAIN) {
			_ = c.Close()
			c.mux.Unlock()
			c.poller.removeConn(c)
			return 0, err
		}
		if n < 0 {
			n = 0
		}
		left := len(b) - n
		if left > 0 {
			c.writeBuffer = byteBuffer.Get(left)
			copy(c.writeBuffer, b[n:])
			_ = c.poller.ModWrite(c.fd)
		}
		c.mux.Unlock()
		return len(b), nil
	}
	c.writeBuffer = append(c.writeBuffer, b...)
	c.mux.Unlock()

	return len(b), nil
}

func (c *conn) Close() error {
	c.closed.Store(true)
	return syscall.Close(c.fd)
}

func (c *conn) LocalAddr() net.Addr {
	return c.localAddr
}

func (c *conn) RemoteAddr() net.Addr {
	return c.rAddr
}

// SetDeadline unSupport
func (c *conn) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline unSupport
func (c *conn) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline unSupport
func (c *conn) SetWriteDeadline(t time.Time) error {
	return nil
}
