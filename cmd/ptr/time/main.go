package main

import (
	"fmt"

	"github.com/something-that-is-cool/zutil/pkg/win"
	"github.com/something-that-is-cool/zutil/pkg/win/mem"
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

	v, finalAddr, err := mem.ResolvePointerValue[int32](proc, baseAddr, offsets)
	if err != nil {
		panic(fmt.Errorf("resolve pointer value: %w", err))
	}
	fmt.Println("resolved pointer value:", v)
	fmt.Printf("address: 0x%x\n", finalAddr)
}
