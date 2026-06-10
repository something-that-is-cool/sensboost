package win

import (
	"path/filepath"

	w "golang.org/x/sys/windows"
)

func ForegroundIs(targetProc ...string) bool {
	pid, ok := ForegroundPID()
	if !ok {
		return false
	}
	h, err := OpenProcessHandle(pid)
	if err != nil {
		return false
	}
	defer w.CloseHandle(h) //nolint:errcheck

	name, ok := ProcessName(h)
	if !ok {
		return false
	}
	for _, target := range targetProc {
		if name == target {
			return true
		}
	}
	return false
}

func ProcessName(h w.Handle) (string, bool) {
	buf := make([]uint16, w.MAX_PATH)
	size := uint32(len(buf))

	if err := w.QueryFullProcessImageName(h, 0, &buf[0], &size); err != nil {
		return "", false
	}
	return filepath.Base(w.UTF16ToString(buf[:size])), true
}

func ForegroundPID() (uint32, bool) {
	var pid uint32
	if _, err := w.GetWindowThreadProcessId(w.GetForegroundWindow(), &pid); err != nil {
		return 0, false
	}
	return pid, true
}

func OpenProcessHandle(pid uint32) (w.Handle, error) {
	h, err := w.OpenProcess(w.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return 0, err
	}
	return h, nil
}
