package asm

import "encoding/binary"

//todo: impl builder

func Jmp(from, to uintptr) []byte {
	offset := int32(to - (from + 5))
	res := make([]byte, 5)
	res[0] = 0xE9
	binary.LittleEndian.PutUint32(res[1:], uint32(offset))
	return res
}

func MovRaxAbs(address uintptr) []byte {
	res := make([]byte, 10)
	res[0] = 0x48
	res[1] = 0xA3
	binary.LittleEndian.PutUint64(res[2:], uint64(address))
	return res
}

func MovEaxAbs(address uintptr) []byte {
	res := make([]byte, 5)
	res[0] = 0xA3
	binary.LittleEndian.PutUint32(res[1:], uint32(address))
	return res
}

func LeaR13(offset uint32) []byte {
	res := make([]byte, 7)
	res[0] = 0x49
	res[1] = 0x8D
	res[2] = 0x85
	binary.LittleEndian.PutUint32(res[3:], offset)
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

func MovssXmm0Rax() []byte {
	return []byte{0xF3, 0x0F, 0x10, 0x00}
}

func MovssXmmToRax(xmmIndex byte) []byte {
	if xmmIndex <= 7 {
		return []byte{0xF3, 0x0F, 0x11, xmmIndex << 3}
	}
	return []byte{0xF3, 0x44, 0x0F, 0x11, (xmmIndex & 7) << 3}
}
