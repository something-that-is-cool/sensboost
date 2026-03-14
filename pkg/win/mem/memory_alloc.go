package mem

import (
	"errors"
	"fmt"
	"syscall"

	"github.com/something-that-is-cool/zutil/pkg/win"
	w "golang.org/x/sys/windows"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procVirtualAllocEx = kernel32.NewProc("VirtualAllocEx")
	procVirtualFreeEx  = kernel32.NewProc("VirtualFreeEx")
)

func Alloc(proc *win.Process, size uintptr) (uintptr, error) {
	ret, _, err := procVirtualAllocEx.Call(
		uintptr(proc.Handle),
		0,
		size,
		uintptr(w.MEM_COMMIT|w.MEM_RESERVE),
		uintptr(w.PAGE_EXECUTE_READWRITE),
	)
	if ret == 0 {
		if err == nil {
			err = errors.New("cannot call proc")
		}
		return 0, fmt.Errorf("call VirtualAllocEx: %w", err)
	}
	return ret, nil
}

func Unalloc(proc *win.Process, addr uintptr) (bool, error) {
	ret, _, err := procVirtualFreeEx.Call(
		uintptr(proc.Handle),
		addr,
		uintptr(0),             // dwSize must be 0 for MEM_RELEASE
		uintptr(w.MEM_RELEASE), // dwFreeType
	)
	if ret == 0 {
		if err == nil {
			err = errors.New("cannot call proc")
		}
		return false, fmt.Errorf("call VirtualFreeEx: %w", err)
	}
	return true, nil
}

func AllocNear(proc *win.Process, target uintptr, size uintptr) (uintptr, error) {
	for offset := uintptr(0x10000); offset < 0x7FFFF000; offset += 0x10000 {
		for _, addr := range []uintptr{target + offset, target - offset} {
			if addr == 0 {
				continue
			}
			ret, _, _ := procVirtualAllocEx.Call(
				uintptr(proc.Handle),
				addr,
				size,
				uintptr(w.MEM_COMMIT|w.MEM_RESERVE),
				uintptr(w.PAGE_EXECUTE_READWRITE),
			)
			if ret != 0 {
				return ret, nil
			}
		}
	}
	return 0, errors.New("failed to find near memory")
}
