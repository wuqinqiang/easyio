//go:build linux || darwin || netbsd || freebsd || openbsd || dragonfly

package easyio

import (
	"errors"
	"fmt"
	"net"
	"syscall"
)

func dupStdConn(netconn net.Conn) (Conn, error) {
	sc, ok := netconn.(interface {
		SyscallConn() (syscall.RawConn, error)
	})
	if !ok {
		return nil, errors.New("RawConn Unsupported")
	}
	rc, err := sc.SyscallConn()
	if err != nil {
		return nil, errors.New("RawConn Unsupported")
	}

	var newFd int
	errCtrl := rc.Control(func(fd uintptr) {
		fmt.Println("before conn fd:", fd)
		newFd, err = syscall.Dup(int(fd))
	})
	if errCtrl != nil {
		return nil, errCtrl
	}

	if err != nil {
		return nil, err
	}

	fmt.Println("after conn fd:", newFd)

	lAddr := netconn.LocalAddr()
	rAddr := netconn.RemoteAddr()

	netconn.Close()

	c := &conn{
		fd:        newFd,
		localAddr: lAddr,
		rAddr:     rAddr,
	}

	return c, nil
}
