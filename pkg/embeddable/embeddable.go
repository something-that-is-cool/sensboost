package embeddable

import "fmt"

// Extract tries to extract data from serialized data, allowing to provide
// custom Marshaller.
func Extract[T any](m Marshaller, d []byte) (x T, err error) {
	err = ExtractTo(m, d, &x)
	return x, err
}

func MustExtract[T any](m Marshaller, d []byte) (x T) {
	MustExtractTo(m, d, &x)
	return x
}

func ExtractTo[T any](m Marshaller, d []byte, x *T) error {
	if err := m.Unmarshal(d, x); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	return nil
}

func MustExtractTo[T any](m Marshaller, d []byte, x *T) {
	if err := ExtractTo(m, d, x); err != nil {
		panic(fmt.Errorf("embeddable: MustExtractTo: %w", err))
	}
}
