// Copyright (c) 2023 Remember
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package easyio

import "net"

type Option func(options *Options)

type Options struct {
	numPoller    int
	listener     func(network, addr string) (net.Listener, error) // Listener for accept conns
	eventHandler EventHandler
	byteBuffer   ByteBuffer
}

func WithNumPoller(num int) Option {
	return func(options *Options) {
		options.numPoller = num
	}
}

func WithListener(fn func(network, addr string) (net.Listener, error)) Option {
	return func(options *Options) {
		options.listener = fn
	}
}

func WithEventHandler(handler EventHandler) Option {
	return func(options *Options) {
		options.eventHandler = handler
	}
}

func WithByteBuffer(byteBuffer ByteBuffer) Option {
	return func(options *Options) {
		options.byteBuffer = byteBuffer
	}
}
