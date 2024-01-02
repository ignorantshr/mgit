package util

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
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

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	out.Close()
	return err
}

func CopySymLink(src, dst string) error {
	link, err := os.Readlink(src)
	if err != nil {
		return err
	}
	return os.Symlink(link, dst)
}

func CopyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relp, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		newPath := filepath.Join(dst, relp)

		if d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			return os.MkdirAll(newPath, info.Mode())
		} else if d.Type() == os.ModeSymlink {
			return CopySymLink(path, newPath)
		} else {
			return CopyFile(path, newPath)
		}
	})
}
