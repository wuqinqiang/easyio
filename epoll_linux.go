// Copyright (c) 2023 Remember
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build linux || darwin || netbsd || freebsd || openbsd || dragonfly

package easyio

import (
	"errors"
	"runtime"
	"sync/atomic"
	"syscall"
)

type Poller struct {
	e *Engine

	index int

	fd       int
	wfd      int
	shutdown atomic.Bool
}

func NewPoller(e *Engine) (*Poller, error) {
	p := new(Poller)
	p.e = e

	fd, err := syscall.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	p.fd = fd

	r1, _, err0 := syscall.Syscall(syscall.SYS_EVENTFD2, 0, syscall.O_NONBLOCK, 0)
	if err0 != 0 {
		syscall.Close(p.fd) //nolint:err
		return nil, err0
	}
	p.wfd = int(r1)

	if err = syscall.EpollCtl(fd, syscall.EPOLL_CTL_ADD, int(r1),
		&syscall.EpollEvent{Fd: int32(r1),
			Events: syscall.EPOLLIN,
		},
	); err != nil {
		syscall.Close(p.fd)  //nolint:err
		syscall.Close(p.wfd) //nolint:err
		return nil, err
	}

	return p, nil
}

func (p *Poller) Wait() error {
	mesc := -1
	events := make([]syscall.EpollEvent, 1024)

	handler := p.e.options.event

	for !p.shutdown.Load() {
		n, err := syscall.EpollWait(p.fd, events, mesc)
		if err != nil && errors.Is(err, syscall.EINTR) {
			return err
		}
		// no event
		if n <= 0 {
			mesc = -1
			runtime.Gosched()
			continue
		}
		mesc = 20

		for i := 0; i < n; i++ {
			event := events[i]

			//find conn
			c := p.e.GetConn(int(event.Fd))
			if c == nil {
				syscall.Close(int(event.Fd))
				continue
			}

			//写
			if event.Events&WriteEvents != 0 {
				c.Flush()
			}

			//event Error
			if event.Events&ErrEvents != 0 {
				p.closeConn(c)
				continue
			}

			//read event
			if event.Events&ReadEvents != 0 {
				handler.OnRead(c.Context(), c)
			}
		}
	}

	return nil
}

func (p *Poller) closeConn(c Conn) {
	if c == nil {
		return
	}

	c.Close()
	p.removeConn(c)

	e := p.e.options.event
	if e != nil {
		e.OnClose(c.Context(), c)
	}
}

func (p *Poller) removeConn(conn Conn) {
	if conn == nil {
		return
	}

	fd := conn.Fd()
	if c := p.e.GetConn(fd); c != nil {
		p.e.Remove(fd)
		_ = p.Delete(fd)
	}
}

func (p *Poller) Close() error {
	p.shutdown.Store(true)
	syscall.Close(p.wfd) //nolint:errcheck
	return syscall.Close(p.fd)
}

func (p *Poller) AddRead(fd int) error {
	return syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_ADD, fd, &syscall.EpollEvent{Fd: int32(fd),
		Events: syscall.EPOLLERR | syscall.EPOLLHUP | syscall.EPOLLRDHUP | syscall.EPOLLPRI | syscall.EPOLLIN,
	})
}
func (p *Poller) ModWrite(fd int) error {
	return syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_MOD, fd, &syscall.EpollEvent{Fd: int32(fd),
		Events: syscall.EPOLLERR | syscall.EPOLLHUP | syscall.EPOLLRDHUP | syscall.EPOLLPRI | syscall.EPOLLIN | syscall.EPOLLOUT,
	})
}

func (p *Poller) ModRead(fd int) error {
	return syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_MOD, fd, &syscall.EpollEvent{Fd: int32(fd),
		Events: syscall.EPOLLERR | syscall.EPOLLHUP | syscall.EPOLLRDHUP | syscall.EPOLLPRI | syscall.EPOLLIN,
	})
}

func (p *Poller) Delete(fd int) error {
	return syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_DEL, fd, &syscall.EpollEvent{Fd: int32(fd)})
}
