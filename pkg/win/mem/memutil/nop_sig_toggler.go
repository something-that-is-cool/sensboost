package memutil

import (
	"errors"
	"fmt"

	"github.com/something-that-is-cool/zutil/pkg/win"
	"github.com/something-that-is-cool/zutil/pkg/win/mem"
)

type SignatureNopTogglerConfig struct {
	Process   *win.Process
	Signature mem.Signature
	NopSig    mem.Signature
}

func (conf SignatureNopTogglerConfig) New() (*SignatureNopToggler, error) {
	t := &SignatureNopToggler{
		pr:  conf.Process,
		sig: conf.Signature,
		nop: conf.NopSig,
	}
	addr, isPatched, err := t.scanAddress()
	if err != nil {
		return nil, fmt.Errorf("initial sig scan: %w", err)
	}
	t.ByteToggler = ByteToggler{
		Process:  conf.Process,
		Address:  addr,
		Original: t.sig.Data,
		Patch:    mem.NopBytes(len(t.sig.Data)),
	}
	t.ByteToggler.SetState(isPatched)
	return t, nil
}

type SignatureNopToggler struct {
	ByteToggler

	pr       *win.Process
	sig, nop mem.Signature
}

func (t *SignatureNopToggler) Toggle() error {
	return t.ByteToggler.Set(!t.ByteToggler.Enabled())
}

func (t *SignatureNopToggler) scanAddress() (addr uintptr, patched bool, err error) {
	addr, err = mem.ScanSignature(t.pr, t.sig)
	if err == nil {
		return addr, false, nil
	}
	addr, err = mem.ScanSignature(t.pr, t.mustNopSig())
	if err == nil {
		return addr, true, nil
	}
	return 0, false, errors.New("cannot find signature")
}

func (t *SignatureNopToggler) mustNopSig() mem.Signature {
	if t.nop.Mask == "" || len(t.nop.Data) == 0 {
		t.nop = createNopSig(len(t.sig.Data))
		return t.nop
	}
	return t.nop
}

func createNopSig(size int) mem.Signature {
	return mem.Signature{Data: mem.NopBytes(size), Mask: mem.FullMask(size)}
}
