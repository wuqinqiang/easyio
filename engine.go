package easyio

import (
	"fmt"
	"runtime"
)

var (
	MaxOpenFiles = 1024 * 1024 * 2
)

type Option func(options *Options)

type Options struct {
	numPoller int
	connEvent ConnHandler
}

func New(network string, addr string, fns ...Option) *Engine {
	e := new(Engine)
	o := new(Options)
	for _, opt := range fns {
		opt(o)
	}
	e.options = o
	e.exitCh = make(chan struct{})
	e.network = network
	e.addr = addr
	return e
}

type Engine struct {
	network string
	addr    string

	exitCh chan struct{}

	listener     *Listener
	pollerManger *Manger
	conns        []Conn

	options *Options
}

func (e *Engine) Start() (err error) {
	fmt.Println("engine start")
	defer func() {
		if err != nil {
			if e.listener != nil {
				e.listener.Close() //nolint:err
			}

		}
	}()

	e.init()

	e.listener, err = NewListener(e.network, e.addr)
	if err != nil {
		return err
	}

	//init poller manger
	e.pollerManger, err = NewManger(e, e.options.numPoller)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-e.exitCh:
				return
			default:
				nconn, err := e.listener.Accept()
				if err != nil {
					fmt.Println("[listener] Accept:", err)
					//todo log
					continue
				}
				if nconn == nil {
					continue
				}

				ec := nconn.(*conn)
				connfd := ec.Fd()
				fmt.Println("new conn:", connfd)

				poller := e.pollerManger.Pick(connfd)
				if err = poller.AddRead(connfd); err != nil {
					fmt.Println("poller.AddRead:", err)
					nconn.Close()
					return
				}
				e.conns[connfd] = ec
			}
		}
	}()

	return nil
}

func (e *Engine) init() {
	if e.options.numPoller <= 0 {
		e.options.numPoller = runtime.NumCPU()
	}
	if e.options.connEvent == nil {
		e.options.connEvent = new(Default)
	}

	e.conns = make([]Conn, MaxOpenFiles)
}

func (e *Engine) Stop() error {

	// listener close
	e.listener.Close()

	// conns close

	for _, conn := range e.conns {
		if conn == nil {
			continue
		}
		conn.Close()
	}

	// poller stop
	e.pollerManger.Stop()

	return nil
}

func (e *Engine) AddConn(conn Conn) {
	e.conns[conn.Fd()] = conn
}

func (e *Engine) Remove(pd int) {
	e.conns[pd] = nil
}

func (e *Engine) GetConn(pd int) Conn {
	if pd >= len(e.conns) {
		panic("fd conn is not exist")
	}
	return e.conns[pd]
}
