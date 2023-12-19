package util

import (
	"os"
)

func IsFileExist(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func IsDir(p string) bool {
	stat, err := os.Stat(p)
	return err == nil && stat.IsDir()
}

func IsFile(p string) bool {
	stat, err := os.Stat(p)
	return err == nil && !stat.IsDir()
}
