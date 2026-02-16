package sens

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/k4ties/sensboost/pkg/win"
)

type TrackerConfig struct {
	ProcessName string
	BaseOffset  uintptr
	Offsets     []uintptr
	Logger      *slog.Logger
}

// New ...
func (conf TrackerConfig) New() (*Tracker, error) {
	if conf.ProcessName == "" {
		return nil, errors.New("empty process name")
	}
	if conf.BaseOffset <= 0 {
		return nil, errors.New("empty base offset")
	}
	if len(conf.Offsets) == 0 {
		return nil, errors.New("empty offsets")
	}
	proc, err := findProcess(conf.ProcessName)
	if err != nil {
		return nil, fmt.Errorf("find process: %w", err)
	}
	if conf.Logger == nil {
		conf.Logger = slog.Default()
	}
	return &Tracker{proc: proc, conf: conf, close: make(chan struct{})}, nil
}

func findProcess(name string) (*win.Process, error) {
	pid := win.FindPID(name)
	if pid <= 0 {
		return nil, errors.New("no process by name")
	}
	proc, err := win.OpenProcess(pid)
	if err != nil {
		return nil, fmt.Errorf("open process: %w", err)
	}
	return proc, nil
}

type Tracker struct {
	conf TrackerConfig

	handlers []func(float64)

	proc *win.Process

	closed atomic.Bool

	lastValue float64

	wg sync.WaitGroup

	close chan struct{}

	readValueMu sync.RWMutex
	lastRead    time.Time
}

// Run ...
// Caller must close tracker after Run ends.
func (tr *Tracker) Run(ctx context.Context) error {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	tr.wg.Add(1)
	defer tr.wg.Done()

	for {
		select {
		case <-ticker.C:
			if err := tr.readValue(true); err != nil {
				return fmt.Errorf("read value: %w", err)
			}
		case <-tr.close:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (tr *Tracker) ForceRead() error {
	return tr.readValue(false)
}

var ErrTrackerClosed = errors.New("tracker is closed")

// WriteValue ...
func (tr *Tracker) WriteValue(v float64) error {
	if tr.closed.Load() {
		return ErrTrackerClosed
	}
	tr.readValueMu.Lock()
	defer tr.readValueMu.Unlock()

	_, finalAddr, err := tr.resolveAddress()
	if err != nil {
		return fmt.Errorf("resolve address: %w", err)
	}
	if err = win.WriteMemory(tr.proc, finalAddr, float32(v)); err != nil {
		return fmt.Errorf("write memory: %w", err)
	}
	tr.lastValue = v
	for _, h := range tr.handlers {
		h(v)
	}
	return nil
}

func (tr *Tracker) Handle(f func(float64)) {
	tr.handlers = append(tr.handlers, f)
}

// LastValue ...
func (tr *Tracker) LastValue() float64 {
	tr.readValueMu.RLock()
	defer tr.readValueMu.RUnlock()
	return tr.lastValue
}

var ErrTrackerAlreadyClosed = errors.New("tracker already closed")

// Close ...
func (tr *Tracker) Close() error {
	if !tr.closed.CompareAndSwap(false, true) {
		return ErrTrackerAlreadyClosed
	}
	close(tr.close)
	err := tr.proc.Close()

	tr.wg.Wait()
	return err
}

func (tr *Tracker) readValue(log bool) error {
	tr.readValueMu.Lock()
	defer tr.readValueMu.Unlock()

	if time.Since(tr.lastRead) < time.Second {
		// prevent calling too fast
		return nil
	}
	tr.wg.Add(1)
	defer tr.wg.Done()

	val, _, err := tr.resolveAddress()
	if err != nil {
		if log {
			tr.conf.Logger.Error("couldn't resolve address (read value)", "err", err.Error())
		}
		return fmt.Errorf("resolve address: %w", err)
	}
	tr.lastValue = float64(val)
	tr.lastRead = time.Now()

	for _, h := range tr.handlers {
		h(float64(val))
	}
	return nil
}

// todo: integrate this to win package
func (tr *Tracker) resolveAddress() (float32, uintptr, error) {
	base, err := tr.proc.GetModuleBase(tr.conf.ProcessName)
	if err != nil {
		return 0, 0, fmt.Errorf("get module base: %w", err)
	}
	addr, err := win.ReadMemory[uintptr](tr.proc, base+tr.conf.BaseOffset)
	if err != nil {
		return 0, 0, fmt.Errorf("read base offset: %w", err)
	}
	for i := 0; i < len(tr.conf.Offsets)-1; i++ {
		addr, err = win.ReadMemory[uintptr](tr.proc, addr+tr.conf.Offsets[i])
		if err != nil {
			return 0, 0, fmt.Errorf("read offset at step %d: %w", i, err)
		}
	}
	finalAddr := addr + tr.conf.Offsets[len(tr.conf.Offsets)-1]
	val, err := win.ReadMemory[float32](tr.proc, finalAddr)
	if err != nil {
		return 0, 0, fmt.Errorf("read final value: %w", err)
	}
	return val, finalAddr, nil
}
