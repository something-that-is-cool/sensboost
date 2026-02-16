package win

import (
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
