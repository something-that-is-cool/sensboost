package main

import (
	"fmt"

	"github.com/something-that-is-cool/zutil/pkg/win"
	"github.com/something-that-is-cool/zutil/pkg/win/mem"
)

var (
	baseAddr uintptr = 0x01921DF8
	offsets          = []uintptr{0x30, 0x78, 0x28, 0x58, 0x0}
)

func main() {
	proc, err := win.OpenProcess("Minecraft.Windows.exe")
	if err != nil {
		panic(fmt.Errorf("open process: %w", err))
	}
	defer proc.Close() //nolint:errcheck

	v, finalAddr, err := mem.ResolvePointerValue[int64](proc, baseAddr, offsets)
	if err != nil {
		panic(fmt.Errorf("resolve pointer value: %w", err))
	}
	fmt.Println("resolved pointer value:", v)
	fmt.Printf("address: 0x%x\n", finalAddr)
}
