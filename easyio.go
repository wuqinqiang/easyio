package easyio

import "fmt"

type ConnHandler interface {
	OnOpen(c Conn)
	OnClose(c Conn)
	OnRead(c Conn)
	OnData(c Conn, data []byte)
}

var _ ConnHandler = (*Default)(nil)

type Default struct{}

func (d *Default) OnOpen(c Conn) {
	panic("implement me")
}

func (d *Default) OnClose(c Conn) {
	//TODO implement me
	panic("implement me")
}

func (d *Default) OnRead(c Conn) {
	b := make([]byte, 1024) // 分配足够的容量
	n, err := c.Read(b[:cap(b)])
	if err != nil {
		fmt.Println("OnRead err:", err)
	}
	fmt.Printf("data len:%v,data:%v\n", n, string(b[:n]))
}

func (d *Default) OnData(c Conn, data []byte) {
	//TODO implement me
	panic("implement me")
}
