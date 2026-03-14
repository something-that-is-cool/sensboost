// Package mem contains a set of memory utilities.
//
// credits: github.com/Zwuiix-cmd/external-aimassist
package mem

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/something-that-is-cool/zutil/pkg/win"
	w "golang.org/x/sys/windows"
)

func WriteMemory[T any](proc *win.Process, addr uintptr, val T) error {
	size := unsafe.Sizeof(val)

	oldProtect, err := virtualProtectUnlock(proc, addr, size)
	if err != nil {
		return fmt.Errorf("virtual protect (unlock): %w", err)
	}
	defer virtualProtectLock(proc, addr, size, oldProtect)

	err = w.WriteProcessMemory(
		proc.Handle,
		addr,
		(*byte)(unsafe.Pointer(&val)),
		size,
		nil,
	)
	return nil
}

func WriteNop(proc *win.Process, addr uintptr, size uint) error {
	return Patch(proc, addr, NopBytes(int(size)))
}

func ReadMemory[T any](p *win.Process, addr uintptr, protect ...bool) (val T, err error) {
	size := unsafe.Sizeof(val)
	if hasTrueOption(protect) {
		oldProtect, err := virtualProtectUnlock(p, addr, size)
		if err != nil {
			return val, fmt.Errorf("virtual protect (unlock): %w", err)
		}
		defer virtualProtectLock(p, addr, size, oldProtect)
	}
	err = w.ReadProcessMemory(
		p.Handle,
		addr,
		(*byte)(unsafe.Pointer(&val)),
		size,
		nil,
	)
	return val, err
}

func ReadBytes(p *win.Process, addr uintptr, size uint, protect ...bool) ([]byte, error) {
	if size == 0 {
		return nil, errors.New("zero size")
	}
	if hasTrueOption(protect) {
		oldProtect, err := virtualProtectUnlock(p, addr, uintptr(size))
		if err != nil {
			return nil, fmt.Errorf("virtual protect (unlock): %w", err)
		}
		defer virtualProtectLock(p, addr, uintptr(size), oldProtect)
	}
	buf := make([]byte, size)

	var bytesRead uintptr
	if err := w.ReadProcessMemory(p.Handle, addr, &buf[0], uintptr(size), &bytesRead); err != nil {
		return nil, err
	}
	if bytesRead != uintptr(size) {
		return nil, fmt.Errorf("read %d bytes, expected %d", bytesRead, size)
	}
	return buf, nil
}

func ResolvePointerValue[T any](proc *win.Process, baseAddr uintptr, offsets []uintptr) (T, uintptr, error) {
	var zero T
	finalAddr, err := ResolvePointerAddress(proc, baseAddr, offsets)
	if err != nil {
		return zero, 0, fmt.Errorf("resolve pointer address: %w", err)
	}
	val, err := ReadMemory[T](proc, finalAddr)
	if err != nil {
		return zero, 0, fmt.Errorf("read final value: %w", err)
	}
	return val, finalAddr, nil
}

func ResolvePointerAddress(proc *win.Process, baseAddr uintptr, offsets []uintptr) (uintptr, error) {
	mod, _, err := proc.GetModuleInfo()
	if err != nil {
		return 0, fmt.Errorf("get proc module info: %w", err)
	}
	addr, err := ReadMemory[uintptr](proc, mod+baseAddr)
	if err != nil {
		return 0, fmt.Errorf("read base addr: %w", err)
	}
	for i := 0; i < len(offsets)-1; i++ {
		addr, err = ReadMemory[uintptr](proc, addr+offsets[i])
		if err != nil {
			return 0, fmt.Errorf("read offset at step %d: %w", i, err)
		}
	}
	return addr + offsets[len(offsets)-1], nil
}

func Patch(p *win.Process, addr uintptr, b []byte) error {
	if len(b) == 0 {
		return errors.New("empty slice")
	}
	var oldProtect uint32
	err := w.VirtualProtectEx(p.Handle, addr, uintptr(len(b)), w.PAGE_EXECUTE_READWRITE, &oldProtect)
	if err != nil {
		return fmt.Errorf("virtual protect: %w", err)
	}
	err = w.WriteProcessMemory(p.Handle, addr, &b[0], uintptr(len(b)), nil)
	if err != nil {
		return fmt.Errorf("write memory: %w", err)
	}
	return w.VirtualProtectEx(p.Handle, addr, uintptr(len(b)), oldProtect, &oldProtect)
}

const sigChunkSize = 1024 * 1024

func ScanSignature(proc *win.Process, sig Signature) (uintptr, error) {
	modBase, modSize, err := proc.GetModuleInfo()
	if err != nil {
		return 0, fmt.Errorf("get module info: %w", err)
	}
	buffer := make([]byte, sigChunkSize)
	sigLen := len(sig.Data)

	for offset := uintptr(0); offset < modSize; {
		toRead := uintptr(sigChunkSize)
		if offset+sigChunkSize > modSize {
			toRead = modSize - offset
		}
		var bytesRead uintptr
		if err := w.ReadProcessMemory(proc.Handle, modBase+offset, &buffer[0], toRead, &bytesRead); err != nil || bytesRead < uintptr(sigLen) {
			offset += toRead
			continue
		}
		if foundOffset, ok := findInChunk(buffer[:bytesRead], sig); ok {
			return modBase + offset + uintptr(foundOffset), nil
		}
		offset += toRead - uintptr(sigLen) + 1
	}
	return 0, errors.New("signature not found")
}

func NopBytes(size int) []byte {
	res := make([]byte, 0, size)
	for range size {
		res = append(res, 0x90)
	}
	return res
}

func findInChunk(chunk []byte, sig Signature) (int, bool) {
	for i := 0; i <= len(chunk)-len(sig.Data); i++ {
		if matchSignature(chunk[i:i+len(sig.Data)], sig) {
			return i, true
		}
	}
	return 0, false
}

func virtualProtectUnlock(proc *win.Process, addr, size uintptr) (p uint32, err error) {
	err = w.VirtualProtectEx(proc.Handle, addr, size, w.PAGE_EXECUTE_READWRITE, &p)
	if err != nil {
		return 0, fmt.Errorf("virtual protect (unlock): %w", err)
	}
	return p, nil
}

func virtualProtectLock(proc *win.Process, addr, size uintptr, p uint32) {
	var temp uint32
	_ = w.VirtualProtectEx(proc.Handle, addr, size, p, &temp)
}

func hasTrueOption(opts []bool) bool {
	for _, opt := range opts {
		if opt {
			return true
		}
	}
	return false
}
