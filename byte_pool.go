// Copyright (c) 2023 Remember
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package easyio

type ByteBuffer interface {
	Get(size int) []byte
	Put(buf []byte)
}

type BufferDefault struct{}

func (d *BufferDefault) Get(size int) []byte { return make([]byte, size) }
func (d *BufferDefault) Put(_ []byte)        {}
