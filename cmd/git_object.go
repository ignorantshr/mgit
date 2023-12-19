package cmd

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/ignorantshr/mgit/util"
)

type object interface {
	format() string
	serialize(repo *repository) []byte
	deserialize(data []byte)
}

func readObject(repo *repository, sha string) object {
	path, err := repo.repoFile(false, "objects", sha[:2], sha[2:])
	if err != nil {
		util.PanicErr(err)
	}

	if !util.IsFile(path) {
		util.PanicErr(_readObjectErr(nil, "file is not found"))
	}

	f, err := os.ReadFile(path)
	if err != nil {
		util.PanicErr(err)
	}

	r, err := zlib.NewReader(bytes.NewReader(f))
	if err != nil {
		util.PanicErr(err)
	}
	defer r.Close()

	raw, err := io.ReadAll(r)
	if err != nil {
		util.PanicErr(err)
	}

	// Read object type
	i := bytes.IndexByte(raw, ' ')
	if i == -1 {
		util.PanicErr(err)
	}
	format := string(raw[:i])

	raw = raw[i:]
	// Read object size
	i = bytes.IndexByte(raw, '\x00') // null byte
	if i == -1 {
		util.PanicErr(_readObjectErr(err, "format is not correct"))
	}

	size, err := strconv.Atoi(string(raw[:i]))
	if err != nil {
		util.PanicErr(err)
	}
	if size != len(raw)-i-1 {
		util.PanicErr(_readObjectErr(err, "size is not correct"))
	}

	raw = raw[i:]
	var constructor func([]byte) object

	switch format {
	case "commit":
	case "tree":
	case "tag":
	case "blob":
		constructor = newBlob
	default:
		util.PanicErr(_readObjectErr(nil, "unknown format "+format))
	}

	return constructor(raw)
}

func writeObject(obj object, repo *repository) string {
	payload := obj.serialize(repo)

	result := append([]byte(obj.format()), ' ')
	result = strconv.AppendInt(result, int64(len(payload)), 10)
	result = append(result, '\x00')
	result = append(result, payload...)

	rawsha := sha1.Sum(result)
	sha := hex.EncodeToString(rawsha[:])

	if repo != nil {
		p, err := repo.repoFile(true, "objects", string(sha[:2]), string(sha[2:]))
		if err != nil {
			util.PanicErr(err)
		}
		if !util.IsFileExist(p) {
			f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				util.PanicErr(err)
			}
			defer f.Close()

			_, err = f.Write(result)
			if err != nil {
				util.PanicErr(err)
			}
		}
	}

	return sha[:]
}

// 存储用户文件数据，不需要特殊处理
type blob struct {
	fmt  string
	data []byte
}

func newBlob(data []byte) object {
	return &blob{"blob", data}
}

func (b *blob) format() string {
	return b.fmt
}

func (b *blob) serialize(_ *repository) []byte {
	return b.data
}

func (b *blob) deserialize(data []byte) {
	b.data = data
}

func _readObjectErr(err error, reason string) error {
	return errors.Join(err, fmt.Errorf("%s: %s", ERROR_READ_OBJECT, reason))
}

// todo
func findObject(repo *repository, name, format string, follow bool) string {
	return name
}
