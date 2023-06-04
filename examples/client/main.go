package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	conn, err := net.Dial("tcp", ":8090")
	if err != nil {
		panic(err)
	}
	n, err := conn.Write([]byte("hello world"))
	if err != nil {
		panic(err)
	}

	go func() {
		b := make([]byte, 100)
		if n, err = conn.Read(b); err != nil {
			panic(err)
		}
		fmt.Println("read data:", n, string(b))
	}()

	defer conn.Close()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-ch
}
