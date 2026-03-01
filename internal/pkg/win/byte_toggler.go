package win

import (
	"fmt"
	"sync/atomic"
)

type ByteToggler struct {
	Process  *Process
	Address  uintptr
	Offset   uintptr //optional
	Original []byte
	Patch    []byte

	state atomic.Bool
}

func (t *ByteToggler) Set(b bool) error {
	if t.state.Load() == b {
		// prevent redundant calls
		return fmt.Errorf("state is already %t", b)
	}
	data := t.Original
	if b {
		data = t.Patch
	}
	if err := Patch(t.Process, t.Address+t.Offset, data); err != nil {
		return err
	}
	t.state.Store(b)
	return nil
}

func (t *ByteToggler) Enabled() bool {
	return t.state.Load()
}

func (t *ByteToggler) SetState(b bool) {
	t.state.Store(b)
}
