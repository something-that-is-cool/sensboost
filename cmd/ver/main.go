package main

import (
	"fmt"

	"github.com/something-that-is-cool/zutil/pkg/win"
)

func main() {
	proc, err := win.OpenProcess("Minecraft.Windows.exe")
	if err != nil {
		panic(fmt.Errorf("open process: %w", err))
	}
	defer proc.Close() //nolint:errcheck

	fmt.Printf("%#v (%s)\n", proc.Version(), proc.Version())
}
