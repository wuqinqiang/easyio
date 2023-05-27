package easyio

import (
	"fmt"
	"net"
	"runtime"
)

var (
	MaxOpenFiles = 1024 * 1024 * 2
)

type Option func(options *Options)

type Options struct {
	numPoller int
	connEvent ConnHandler
	Listener  func(network, addr string) (net.Listener, error) // Listener for accept conns
}

func New(network, addr string, fns ...Option) *Engine {
	e := new(Engine)
	opts := new(Options)
	for _, opt := range fns {
		opt(opts)
	}

	if opts.Listener == nil {
		opts.Listener = net.Listen
	}

	e.options = opts
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

func (e *Engine) acceptPolling(localOSThread bool) error {
	if localOSThread {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
	}

	for {
		select {
		case <-e.exitCh:
			fmt.Println("engine out")
			return nil
		default:
			nc, err := e.listener.Accept()
			if err != nil {
				fmt.Println("[listener] Accept:", err)
				continue
			}
			if nc == nil {
				continue
			}
			ec := nc.(*conn)
			poller := e.pollerManger.Pick(ec.Fd())
			if err = poller.AddRead(ec.Fd()); err != nil {
				fmt.Println("poller.AddRead:", err)
				nc.Close()
				continue
			}
			e.conns[ec.Fd()] = ec
		}
	}
}

func (e *Engine) Start() (err error) {
	fmt.Println("engine start")
	e.init()
	// new a listener

	listener := new(Listener)
	listener.engine = e

	ln, err := e.options.Listener(e.network, e.addr)
	if err != nil {
		return err
	}
	listener.ln = ln
	listener.addr = ln.Addr()
	e.listener = listener

	//init poller manger
	if e.pollerManger, err = NewManger(e, e.options.numPoller); err != nil {
		return err
	}

	go e.acceptPolling(true) //nolint:errcheck

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
	close(e.exitCh)
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
