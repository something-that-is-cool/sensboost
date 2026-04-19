package asm

import (
	"bytes"
	"strings"
	"sync"

	"github.com/something-that-is-cool/zutil/pkg/win/mem"
)

var builderPool = sync.Pool{ // the pool itself is NOT currently required but when it will be required i will probably forget about it so ill make it now
	New: func() any {
		return new(Builder)
	},
}

type Builder struct {
	buf  bytes.Buffer
	mask strings.Builder
}

func Build() *Builder {
	b := builderPool.Get().(*Builder)
	b.Clear()
	return b
}

func (b *Builder) write(data ...byte) {
	b.buf.Write(data)
	for range len(data) {
		b.mask.WriteByte('x')
	}
}

func (b *Builder) Result() []byte {
	return b.buf.Bytes()
}

func (b *Builder) BuildSignature() mem.Signature {
	sig := mem.Signature{
		Data: b.buf.Bytes(),
		Mask: b.mask.String(),
	}
	return sig
}

func (b *Builder) ZeroByte() *Builder {
	return b.Raw(0x00)
}

func (b *Builder) X() *Builder {
	b.buf.WriteByte(0x00)
	b.mask.WriteByte('?')
	return b
}

func (b *Builder) Raw(v ...byte) *Builder {
	b.write(v...)
	return b
}

func (b *Builder) MovAxImm8(val byte) *Builder {
	return b.Raw(MovAxImm8(val)...)
}

func (b *Builder) XorChBpl() *Builder {
	return b.Raw(XorChBpl()...)
}

func (b *Builder) Jmp(from, to uintptr) *Builder {
	return b.Raw(Jmp(from, to)...)
}

func (b *Builder) MovRax64(address uintptr) *Builder {
	return b.Raw(MovRax64(address)...)
}

func (b *Builder) PushRax() *Builder {
	return b.Raw(PushRax())
}

func (b *Builder) PopRax() *Builder {
	return b.Raw(PopRax())
}

func (b *Builder) Pushfq() *Builder {
	return b.Raw(Pushfq())
}

func (b *Builder) Popfq() *Builder {
	return b.Raw(Popfq())
}

func (b *Builder) Ret() *Builder {
	return b.Raw(Ret())
}

func (b *Builder) MovssXmmToRax(xmmIndex byte) *Builder {
	return b.Raw(MovssXmmToRax(xmmIndex)...)
}

func (b *Builder) MovssXmm0Rax() *Builder {
	return b.Raw(MovssXmm0Rax()...)
}

func (b *Builder) MovRdx(val uintptr) *Builder {
	return b.Raw(MovRdx64(val)...)
}

func (b *Builder) MovMemsdRDX(offset byte, val float32) *Builder {
	return b.Raw(MovMemsdRDX(offset, val)...)
}

func (b *Builder) Float(val float32) *Builder {
	return b.Raw(Float(val)...)
}

func (b *Builder) Clear() *Builder {
	b.buf.Reset()
	b.mask.Reset()
	return b
}

func (b *Builder) ClearAndReturn() {
	b.Clear()
	builderPool.Put(b)
}
