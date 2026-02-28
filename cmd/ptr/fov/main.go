package main

import (
	"fmt"

	"github.com/something-that-is-cool/zutil/internal/pkg/win"
)

var (
	baseAddr uintptr = 0x01921DF8
	offsets          = []uintptr{0x30, 0xD8, 0x20, 0xBE0}
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
	v, finalAddr, err := win.ResolvePointerValue[float32](proc, module, baseAddr, offsets)
	if err != nil {
		panic(fmt.Errorf("resolve pointer value: %w", err))
	}
	_ = finalAddr // do something
	fmt.Println("resolved float pointer value:", v)
}
