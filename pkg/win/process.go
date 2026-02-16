package win

import (
	"errors"
	"unsafe"

	"github.com/go-vgo/robotgo"
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
}

func OpenProcess(pid uint32) (*Process, error) {
	h, err := w.OpenProcess(w.PROCESS_ALL_ACCESS, false, pid)
	if err != nil {
		return nil, err
	}
	return &Process{PID: pid, Handle: h}, nil
}

func (proc *Process) Close() error {
	return w.CloseHandle(proc.Handle)
}

func (proc *Process) GetModuleBase(moduleName string) (uintptr, error) {
	snapshot, err := w.CreateToolhelp32Snapshot(w.TH32CS_SNAPMODULE|w.TH32CS_SNAPMODULE32, proc.PID)
	if err != nil {
		return 0, err
	}
	defer w.CloseHandle(snapshot) //nolint:errcheck

	var me w.ModuleEntry32
	me.Size = uint32(unsafe.Sizeof(me))

	if err = w.Module32First(snapshot, &me); err != nil {
		return 0, err
	}
	for {
		if w.UTF16ToString(me.Module[:]) == moduleName {
			return me.ModBaseAddr, nil
		}
		if err = w.Module32Next(snapshot, &me); err != nil {
			break
		}
	}
	return 0, errors.New("no such module")
}
