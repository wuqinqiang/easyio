// Copyright (c) 2023 Remember
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build linux

package easyio

import (
	"errors"
	"fmt"
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
	fmt.Println("event fd:", r1)
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

	handler := p.e.options.connEvent

	for !p.shutdown.Load() {
		n, err := syscall.EpollWait(p.fd, events, mesc)
		if err != nil && errors.Is(err, syscall.EINTR) {
			return err
		}
		// no event
		if n <= 0 {
			mesc = -1
			continue
		}
		mesc = 20

		for i := 0; i < n; i++ {
			event := events[i]
			fmt.Printf("event:%+v\n", event)
			//find conn
			conn := p.e.GetConn(int(event.Fd))
			if conn == nil {
				syscall.Close(int(event.Fd))
				continue
			}
			// closed event
			if event.Events&(syscall.EPOLLHUP|syscall.EPOLLRDHUP) != 0 {
				//remove the conn
				p.e.Remove(conn.Fd())
				// delete  events
				p.Delete(conn.Fd())
				//close the fd
				conn.Close()
				continue
			}

			//event Error
			if event.Events&syscall.EPOLLERR != 0 {
				fmt.Println("EPOLLERR error:", event)
				continue
			}

			//判断是哪种事件
			if event.Events&(syscall.EPOLLERR|syscall.EPOLLHUP|syscall.EPOLLRDHUP) != 0 {
				conn.Close()
				p.DeleteConn(conn)
				continue
			}
			//写
			if event.Events&syscall.EPOLLOUT != 0 {
				conn.Flush()
			}
			//读
			fmt.Println("event conn fd:", conn.Fd())
			if event.Events&(syscall.EPOLLPRI|syscall.EPOLLIN) != 0 {
				handler.OnRead(conn)
			}
		}
	}

	return nil
}

func (p *Poller) DeleteConn(conn Conn) {
	if conn == nil {
		return
	}

	fd := conn.Fd()
	if p.e.GetConn(fd) == conn {
		p.e.Remove(fd)
		p.Delete(fd)
	}
}

func (p *Poller) Stop() error {
	p.shutdown.Store(true)
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
