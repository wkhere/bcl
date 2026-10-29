//go:build unix

package main

import (
	"os"
	"syscall"
)

// fileStat is the part of the file status needed to compare files.
// On unix it comes from fstat(2) directly: os.File.Stat would return
// os.FileInfo, which makes the linker keep much of the time package.
type fileStat struct {
	st syscall.Stat_t
}

func statFile(f *os.File) (s fileStat, err error) {
	rc, err := f.SyscallConn()
	if err != nil {
		return s, err
	}
	cerr := rc.Control(func(fd uintptr) {
		err = syscall.Fstat(int(fd), &s.st)
	})
	if cerr != nil {
		return s, cerr
	}
	if err != nil {
		return s, &os.PathError{Op: "fstat", Path: f.Name(), Err: err}
	}
	return s, nil
}

func (s fileStat) sameFile(t fileStat) bool {
	return s.st.Dev == t.st.Dev && s.st.Ino == t.st.Ino
}

func (s fileStat) isRegular() bool {
	return s.st.Mode&syscall.S_IFMT == syscall.S_IFREG
}
