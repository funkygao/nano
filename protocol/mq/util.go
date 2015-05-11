package mq

import (
	"os"
)

func atomicRename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}
