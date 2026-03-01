package misc

import "sync"

type ValueWithMutex[T any] struct {
	sync.Mutex
	V T
}
