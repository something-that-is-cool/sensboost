package win

import (
	"errors"
	"fmt"
	"sync"

	"github.com/go-vgo/robotgo"
	"github.com/something-that-is-cool/zutil/internal/misc"
	w "golang.org/x/sys/windows"
)

func FindPID(name string, caseInsensitive ...bool) uint32 {
	processes, err := robotgo.Process()
	if err != nil {
		return 0
	}
	for _, proc := range processes {
		if misc.CompareString(proc.Name, name, caseInsensitive...) {
			return uint32(proc.Pid)
		}
	}
	return 0
}

type Process struct {
	PID    uint32
	Handle w.Handle
	Name   string
	Path   string

	ver ProcessVersion

	modules   misc.CaseInsensitiveMap[ProcessModule]
	modulesMu sync.RWMutex
}

func OpenProcess(name string, notLoadModule ...bool) (*Process, error) {
	pid := FindPID(name)
	if pid <= 0 {
		return nil, errors.New("no process by name")
	}
	h, err := w.OpenProcess(w.PROCESS_ALL_ACCESS|w.SYNCHRONIZE, false, pid)
	if err != nil {
		return nil, err
	}
	path, err := ProcessPathByHandle(h)
	if err != nil {
		return nil, fmt.Errorf("get process path: %w", err)
	}
	ver, err := GetProcessVersion(path)
	if err != nil {
		return nil, fmt.Errorf("get process version: %w", err)
	}
	proc := &Process{
		Handle:  h,
		Name:    name,
		PID:     pid,
		Path:    path,
		ver:     ver,
		modules: make(misc.CaseInsensitiveMap[ProcessModule]),
	}
	if !misc.HasTrueOption(notLoadModule) {
		mod, err := proc.GetModuleInfo()
		if err != nil {
			return nil, fmt.Errorf("get module info: %w", err)
		}
		proc.modules.Set(proc.Name, mod)
	}
	return proc, nil
}

func (proc *Process) Close() error {
	return w.CloseHandle(proc.Handle)
}
