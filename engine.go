package easyio

import (
	"fmt"
	"net"
	"runtime"
)

var (
	MaxOpenFiles = 1024 * 1024 * 2
)

func New(network, addr string, fns ...Option) *Engine {
	e := new(Engine)
	opts := new(Options)
	for _, opt := range fns {
		opt(opts)
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

func (e *Engine) Start() (err error) {
	fmt.Println("engine start")
	e.init()
	// new a listener
	ln, err := e.options.Listener(e.network, e.addr)
	if err != nil {
		return err
	}

	listener := new(Listener)
	listener.engine = e
	listener.listener = ln
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

	if e.options.Listener == nil {
		e.options.Listener = net.Listen
	}

	if e.options.numPoller <= 0 {
		e.options.numPoller = runtime.NumCPU()
	}
	if e.options.eventHandler == nil {
		e.options.eventHandler = new(eventHandler)
	}
	if e.options.byteBuffer == nil {
		e.options.byteBuffer = new(BufferDefault)
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

func (e *Engine) GetByteBuffer() ByteBuffer {
	return e.options.byteBuffer
}
func (e *Engine) GetEventHandler() EventHandler {
	return e.options.eventHandler
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

func (e *Engine) acceptPolling(localOSThread bool) error {
	if localOSThread {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
	}

	handler := e.GetEventHandler()

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
			ec.poller = poller

			// set ctx
			ec.ctx = handler.OnOpen(ec)
			if err = poller.AddRead(ec.Fd()); err != nil {
				fmt.Println("poller.AddRead:", err)
				nc.Close()
				continue
			}
			e.conns[ec.Fd()] = ec
		}
	}
}
