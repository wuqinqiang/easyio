package easyio

type ByteBuffer interface {
	Get(size int) []byte
	Put(buf []byte)
}

type BufferDefault struct{}

func (d *BufferDefault) Get(size int) []byte { return make([]byte, size) }
func (d *BufferDefault) Put(_ []byte)        {}
