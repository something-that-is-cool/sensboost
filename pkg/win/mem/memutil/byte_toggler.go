package memutil

import (
	"fmt"
	"sync/atomic"

	"github.com/something-that-is-cool/zutil/pkg/win"
	"github.com/something-that-is-cool/zutil/pkg/win/mem"
)

type ByteToggler struct {
	Process  *win.Process
	Address  uintptr
	Offset   uintptr //optional
	Original []byte
	Patch    []byte

	state atomic.Bool
}

func (t *ByteToggler) Set(b bool) error {
	if t.state.Load() == b {
		return nil
	}
	data := t.Original
	if b {
		data = t.Patch
	}
	targetAddr := t.Address + t.Offset
	if err := mem.Patch(t.Process, targetAddr, data); err != nil {
		return fmt.Errorf("patch at 0x%X: %w", targetAddr, err)
	}
	t.state.Store(b)
	return nil
}

func (t *ByteToggler) Toggle() error {
	return t.Set(!t.Enabled())
}

func (t *ByteToggler) Enabled() bool {
	return t.state.Load()
}

func (t *ByteToggler) SetState(b bool) {
	t.state.Store(b)
}
