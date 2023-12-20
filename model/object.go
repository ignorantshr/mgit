package model

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

type Object interface {
	Format() string
	Serialize(repo *Repository) []byte
	Deserialize(data []byte)
}

func ReadObject(repo *Repository, sha string) Object {
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

	var obj Object
	switch format {
	case "commit":
		obj = NewCommitObj(raw)
	case "tree":
	case "tag":
	case "blob":
		obj = NewBlobObj(raw)
	default:
		util.PanicErr(_readObjectErr(nil, "unknown format "+format))
	}

	return obj
}

func WriteObject(obj Object, repo *Repository) string {
	payload := obj.Serialize(repo)

	result := append([]byte(obj.Format()), ' ')
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

func _readObjectErr(err error, reason string) error {
	return errors.Join(err, fmt.Errorf("%s: %s", ERROR_READ_OBJECT, reason))
}

// todo
func FindObject(repo *Repository, name, format string, follow bool) string {
	return name
}
