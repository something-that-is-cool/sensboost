package mem

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/something-that-is-cool/zutil/pkg/win"
)

type Pointer struct {
	BaseAddress uintptr
	Offsets     []uintptr
}

func ParsePointer(addrStr, offsetsStr string) (ptr Pointer, err error) {
	if !strings.HasPrefix(addrStr, "0x") {
		addrStr = "0x" + addrStr
	}
	i, err := strconv.ParseUint(addrStr, 0, 64)
	if err != nil {
		return Pointer{}, fmt.Errorf("parse addr: %w", err)
	}
	ptr.BaseAddress = uintptr(i)
	for idx, part := range strings.Fields(offsetsStr) {
		if !strings.HasPrefix(part, "0x") {
			part = "0x" + part
		}
		offset, err := strconv.ParseUint(part, 0, 64)
		if err != nil {
			return Pointer{}, fmt.Errorf("parse offset %d (%s): %w", idx, part, err)
		}
		ptr.Offsets = append(ptr.Offsets, uintptr(offset))
	}
	return ptr, nil
}

func MustParsePointer(addrStr, offsetsStr string) Pointer {
	ptr, err := ParsePointer(addrStr, offsetsStr)
	if err != nil {
		panic(fmt.Errorf("mem: MustParsePointer: %w", err))
	}
	return ptr
}

func ResolvePointerValue[T any](proc *win.Process, ptr Pointer, opts ...any) (T, uintptr, error) {
	var zero T
	finalAddr, err := ResolvePointerAddress(proc, ptr, opts...)
	if err != nil {
		return zero, 0, fmt.Errorf("resolve pointer address: %w", err)
	}
	val, err := ReadMemory[T](proc, finalAddr, opts...)
	if err != nil {
		return zero, 0, fmt.Errorf("read final value: %w", err)
	}
	return val, finalAddr, nil
}

func ResolvePointerAddress(proc *win.Process, ptr Pointer, opts ...any) (uintptr, error) {
	mod, err := proc.GetModuleInfo()
	if err != nil {
		return 0, fmt.Errorf("get proc module info: %w", err)
	}
	finalAddr := ptr.BaseAddress
	if subModule, _ := handleOptions(opts); !subModule {
		finalAddr += mod.Address
	}
	addr, err := ReadMemory[uintptr](proc, finalAddr, opts...)
	if err != nil {
		return 0, fmt.Errorf("read base addr: %w", err)
	}
	for i := 0; i < len(ptr.Offsets)-1; i++ {
		addr, err = ReadMemory[uintptr](proc, addr+ptr.Offsets[i], opts...)
		if err != nil {
			return 0, fmt.Errorf("read offset at step %d: %w", i, err)
		}
	}
	return addr + ptr.Offsets[len(ptr.Offsets)-1], nil
}
