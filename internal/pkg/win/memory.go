package win

import (
	"errors"
	"fmt"
	"unsafe"

	w "golang.org/x/sys/windows"
)

func WriteMemory[T any](p *Process, addr uintptr, val T) error {
	return w.WriteProcessMemory(
		p.Handle,
		addr,
		(*byte)(unsafe.Pointer(&val)),
		unsafe.Sizeof(val),
		nil,
	)
}

func ReadMemory[T any](p *Process, addr uintptr) (val T, err error) {
	err = w.ReadProcessMemory(
		p.Handle,
		addr,
		(*byte)(unsafe.Pointer(&val)),
		unsafe.Sizeof(val),
		nil,
	)
	return val, err
}

func ResolvePointerValue[T any](proc *Process, mod, baseAddr uintptr, offsets []uintptr) (T, uintptr, error) {
	var zero T
	finalAddr, err := ResolvePointerAddress(proc, mod, baseAddr, offsets)
	if err != nil {
		return zero, 0, fmt.Errorf("resolve pointer address: %w", err)
	}
	val, err := ReadMemory[T](proc, finalAddr)
	if err != nil {
		return zero, 0, fmt.Errorf("read final value: %w", err)
	}
	return val, finalAddr, nil
}

func ResolvePointerAddress(proc *Process, mod, baseAddr uintptr, offsets []uintptr) (uintptr, error) {
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

func Patch(p *Process, addr uintptr, b []byte) error {
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

func ScanSignature(p *Process, size, base uintptr, pattern []byte) (uintptr, error) {
	if len(pattern) == 0 {
		return 0, errors.New("empty pattern")
	}
	moduleData := make([]byte, size)
	var bytesRead uintptr

	err := w.ReadProcessMemory(p.Handle, base, &moduleData[0], size, &bytesRead)
	if err != nil && bytesRead == 0 {
		return 0, fmt.Errorf("read: %w", err)
	}
	for i := 0; i < int(bytesRead)-len(pattern); i++ {
		match := true
		for j := 0; j < len(pattern); j++ {
			if moduleData[i+j] != pattern[j] {
				match = false
				break
			}
		}
		if match {
			return base + uintptr(i), nil
		}
	}
	return 0, errors.New("signature not found in module memory")
}
