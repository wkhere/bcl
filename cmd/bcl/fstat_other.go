//go:build !unix

package main

import "os"

// fileStat is the part of the file status needed to compare files
// and check their type.
type fileStat struct {
	fi os.FileInfo
}

func statFile(f *os.File) (fileStat, error) {
	fi, err := f.Stat()
	return fileStat{fi}, err
}

// lstatPath doesn't follow a symlink, so it describes what os.Remove
// would remove.
func lstatPath(name string) (fileStat, error) {
	fi, err := os.Lstat(name)
	return fileStat{fi}, err
}

func (s fileStat) sameFile(t fileStat) bool {
	return os.SameFile(s.fi, t.fi)
}

func (s fileStat) isRegular() bool {
	return s.fi.Mode().IsRegular()
}
