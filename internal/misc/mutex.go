package misc

import "sync"

type ValueWithMutex[T any] struct {
	sync.Mutex
	V T
}

type ValueWithRWMutex[T any] struct {
	sync.RWMutex
	V T
}
