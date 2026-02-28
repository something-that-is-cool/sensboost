package main

import (
	"fmt"

	"github.com/something-that-is-cool/zutil/internal/pkg/win"
)

var (
	baseAddr uintptr = 0x01921888
	offsets          = []uintptr{0x10, 0x58, 0x58, 0xD8, 0x0, 0x38, 0x194}
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
	v, finalAddr, err := win.ResolvePointerValue[int32](proc, module, baseAddr, offsets)
	if err != nil {
		panic(fmt.Errorf("resolve pointer value: %w", err))
	}
	fmt.Println("resolved pointer value:", v)
	fmt.Printf("address: 0x%x\n", finalAddr)
}
