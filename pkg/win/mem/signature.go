package mem

import (
	"encoding/hex"
	"fmt"
	"strings"
)

type Signature struct {
	Data []byte
	Mask string // "xxx??x"
}

func MustParseSignature(from string) (sig Signature) {
	for _, part := range strings.Fields(from) {
		if strings.Contains(part, "?") {
			sig.Data = append(sig.Data, 0)
			sig.Mask += "?"
			continue
		}
		if len(part) == 1 {
			part = "0" + part
		}
		b, err := hex.DecodeString(part)
		if err != nil {
			panic(fmt.Errorf("invalid hex byte %q: %w", part, err))
		}
		sig.Data = append(sig.Data, b...)
		sig.Mask += "x"
	}
	return sig
}

func FullMask(n int) string {
	return strings.Repeat("x", n)
}

func matchSignature(data []byte, sig Signature) bool {
	for i := 0; i < len(sig.Data); i++ {
		if sig.Mask[i] == 'x' && data[i] != sig.Data[i] {
			return false
		}
	}
	return true
}
