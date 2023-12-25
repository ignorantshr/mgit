package util

import (
	"os"
)

func IsFileExist(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func IsFile(p string) bool {
	stat, err := os.Stat(p)
	return err == nil && !stat.IsDir()
}

func IsDir(p string) bool {
	stat, err := os.Stat(p)
	return err == nil && stat.IsDir()
}

func IsDirEmpty(p string) bool {
	entries, err := os.ReadDir(p)
	return err == nil && len(entries) == 0
}
