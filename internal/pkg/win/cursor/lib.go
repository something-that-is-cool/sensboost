package cursor

import (
	"errors"
	"syscall"
	"unsafe"

	"github.com/tailscale/win"
)

var (
	user32            = syscall.NewLazyDLL("user32.dll")
	procGetCursorInfo = user32.NewProc("GetCursorInfo")
)

type info struct {
	CbSize uint32
	Flags  uint32

	_ uintptr
	_ win.POINT
}

func getInfo() (i info, err error) {
	i.CbSize = uint32(unsafe.Sizeof(i))
	ret, _, err := procGetCursorInfo.Call(uintptr(unsafe.Pointer(&i)))
	if ret == 0 {
		if err == nil {
			err = errors.New("cannot get cursor info")
		}
		return i, err
	}
	// ...
	return i, nil
}
