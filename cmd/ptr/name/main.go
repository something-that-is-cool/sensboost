package main

import (
	"fmt"

	"github.com/something-that-is-cool/zutil/internal/pkg/win"
)

var (
	baseAddr uintptr = 0x019209F0
	offsets          = []uintptr{0x0, 0x8, 0x0, 0x28, 0x90}
)

func main() {
	proc, err := win.OpenProcess("Minecraft.Windows.exe")
	if err != nil {
		panic(fmt.Errorf("open process: %w", err))
	}
	defer proc.Close() //nolint:errcheck

	module, _, err := proc.GetModuleInfo()
	if err != nil {
		panic(fmt.Errorf("get process module: %w", err))
	}
	v, finalAddr, err := win.ResolvePointerValue[[200]byte](proc, module, baseAddr, offsets)
	if err != nil {
		panic(fmt.Errorf("resolve pointer value: %w", err))
	}
	fmt.Printf("resolved float pointer value: %q\n", toStr(v))
	fmt.Printf("address: 0x%x\n", finalAddr)
}

func toStr(x [200]byte) (s string) {
	for _, v := range x[:] {
		if v == 0 {
			break
		}
		s += string(v)
	}
	return
}
