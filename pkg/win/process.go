package win

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"unsafe"

	"github.com/go-vgo/robotgo"
	"github.com/something-that-is-cool/zutil/internal/misc"
	w "golang.org/x/sys/windows"
)

func FindPID(name string) uint32 {
	processes, err := robotgo.Process()
	if err != nil {
		return 0
	}
	for _, proc := range processes {
		if proc.Name == name {
			return uint32(proc.Pid)
		}
	}
	return 0
}

type Process struct {
	PID    uint32
	Handle w.Handle
	Name   string

	modules   map[string]ProcessModule
	modulesMu sync.RWMutex
}

type ProcessModule struct {
	Address uintptr
	Size    uintptr
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
	proc := &Process{
		Name:    name,
		PID:     pid,
		Handle:  h,
		modules: make(map[string]ProcessModule),
	}
	if !misc.HasTrueOption(notLoadModule) {
		mod, err := proc.GetModuleInfo()
		if err != nil {
			return nil, fmt.Errorf("get module info: %w", err)
		}
		proc.modules[strings.ToLower(proc.Name)] = mod
	}
	return proc, nil
}

type GetModuleInfoOptions struct {
	Name string
	// some future options
}

func (proc *Process) GetModuleInfo(opts ...GetModuleInfoOptions) (ProcessModule, error) {
	if !proc.Active() {
		return ProcessModule{}, errors.New("process is not active")
	}
	opt := proc.extractGetModuleOptions(opts)

	proc.modulesMu.Lock()
	defer proc.modulesMu.Unlock()

	if m, ok := proc.modules[opt.Name]; ok {
		return m, nil
	}
	snapshot, err := w.CreateToolhelp32Snapshot(w.TH32CS_SNAPMODULE|w.TH32CS_SNAPMODULE32, proc.PID)
	if err != nil {
		return ProcessModule{}, err
	}
	defer w.CloseHandle(snapshot) //nolint:errcheck

	var me w.ModuleEntry32
	me.Size = uint32(unsafe.Sizeof(me))

	if err = w.Module32First(snapshot, &me); err != nil {
		return ProcessModule{}, fmt.Errorf("start iterating (m32 first): %w", err)
	}
	for {
		currentModule := w.UTF16ToString(me.Module[:])
		if strings.ToLower(currentModule) == opt.Name {
			mod := ProcessModule{Address: me.ModBaseAddr, Size: uintptr(me.ModBaseSize)}
			proc.modules[opt.Name] = mod
			return mod, nil
		}
		if err = w.Module32Next(snapshot, &me); err != nil {
			break
		}
	}
	return ProcessModule{}, fmt.Errorf("module %q not found", opt.Name)
}

func (proc *Process) Active() bool {
	var exitCode uint32
	if err := w.GetExitCodeProcess(proc.Handle, &exitCode); err != nil {
		return false
	}
	return exitCode == 259 //STILL_ACTIVE
}

func (proc *Process) Close() error {
	return w.CloseHandle(proc.Handle)
}

func (proc *Process) extractGetModuleOptions(x []GetModuleInfoOptions) GetModuleInfoOptions {
	if len(x) == 0 {
		x = make([]GetModuleInfoOptions, 1) //initializes 1 empty opts struct
	}
	opt := x[0]
	if opt.Name == "" {
		opt.Name = proc.Name
	}
	opt.Name = strings.ToLower(opt.Name)
	return opt
}
