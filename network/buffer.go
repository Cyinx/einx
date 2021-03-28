package network

import (
	"sync"
)

const BUFFER_BASE_SIZE = 1024

var bufferPool *sync.Pool = &sync.Pool{New: func() interface{} { return newBytesBuffer() }}

type BytesBuffer struct {
	w   int
	r   int
	c   int
	buf []byte
}

func newBytesBuffer() *BytesBuffer {
	b := &BytesBuffer{
		buf: make([]byte, BUFFER_BASE_SIZE),
	}
	b.c = len(b.buf)
	return b
}

func (b *BytesBuffer) Reset() {
	b.w = 0
	b.r = 0
	b.c = len(b.buf)
}

func (b *BytesBuffer) WriteBuf() []byte {
	return b.buf[b.w:]
}

func (b *BytesBuffer) WriteUint8(m byte) int {
	b.buf[b.w] = m
	return b.write(1)
}

func (b *BytesBuffer) WriteUint32(i uint32) int {
	bigEndian.PutUint32(b.buf[b.w:], i)
	return b.write(4)
}

func (b *BytesBuffer) WriteBytes(m []byte) int {
	copy(b.buf[b.w:], m)
	return b.write(len(m))
}

func (b *BytesBuffer) Write(n int) {
	b.write(n)
}

func (b *BytesBuffer) write(n int) int {
	b.w += n
	return n
}

func (b *BytesBuffer) ReadBuf(n int) []byte {
	if s := b.Count(); s < n {
		n = s
	}
	x := b.buf[b.r : b.r+n]
	b.Read(n)
	return x
}

func (b *BytesBuffer) Read(n int) {
	b.r += n
}

func (b *BytesBuffer) Count() int {
	return b.w - b.r
}

func (b *BytesBuffer) Reserve(s int) {
	if space := (b.c - b.w) + b.r; space < s {
		buf := b.buf
		b.c = b.c + (s - space)
		b.buf = make([]byte, b.c)
		if b.Count() > 0 {
			copy(b.buf[0:], buf[b.r:b.w])
		}
	} else {
		if b.Count() > 0 {
			copy(b.buf[0:], b.buf[b.r:b.w])
		}
	}
	b.w -= b.r
	b.r = 0
}
