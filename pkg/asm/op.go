package asm

import "encoding/binary"

func Jmp(from, to uintptr) []byte {
	offset := int32(to - (from + 5))
	res := make([]byte, 5)
	res[0] = 0xE9
	binary.LittleEndian.PutUint32(res[1:], uint32(offset))
	return res
}

func MovRax64(address uintptr) []byte {
	res := make([]byte, 10)
	res[0] = 0x48
	res[1] = 0xB8
	binary.LittleEndian.PutUint64(res[2:], uint64(address))
	return res
}

func PushRax() byte {
	return 0x50
}

func PopRax() byte {
	return 0x58
}

func Pushfq() byte {
	return 0x9C
}

func Popfq() byte {
	return 0x9D
}

func Ret() byte {
	return 0xC3
}

func MovAxImm8(val byte) []byte {
	return []byte{0x66, 0xB8, val}
}

func XorChBpl() []byte {
	return []byte{0x40, 0x30, 0xED}
}

func MovssXmmToRax(xmm byte) []byte {
	if xmm <= 7 {
		return []byte{0xF3, 0x0F, 0x11, xmm << 3}
	}
	return []byte{0xF3, 0x44, 0x0F, 0x11, (xmm & 7) << 3}
}

func MovssXmm0Rax() []byte {
	return []byte{0xF3, 0x0F, 0x10, 0x00}
}
