package model

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/ignorantshr/mgit/util"
)

type Object interface {
	Format() string
	Serialize(repo *Repository) []byte
	Deserialize(data []byte)
}

func ReadObject(repo *Repository, sha string) Object {
	if sha == "" {
		return nil
	}
	path, err := repo.RepoFile(false, "objects", sha[:2], sha[2:])
	if err != nil {
		util.PanicErr(err)
	}

	if !util.IsFile(path) {
		util.PanicErr(_readObjectErr(nil, "file is not found"))
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		util.PanicErr(err)
	}

	// Read object type
	i := bytes.IndexByte(raw, ' ')
	if i == -1 {
		util.PanicErr(err)
	}
	format := string(raw[:i])

	raw = raw[i+1:]
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

	raw = raw[i+1:]

	var obj Object
	switch format {
	case "commit":
		obj = NewCommitObj()
	case "tree":
		obj = NewTreeObj()
	case "tag":
		obj = NewTagObj()
	case "blob":
		obj = NewBlobObj()
	default:
		util.PanicErr(_readObjectErr(nil, "unknown format "+format))
	}

	obj.Deserialize(raw)
	return obj
}

func WriteObject(repo *Repository, obj Object) string {
	payload := obj.Serialize(repo)

	result := append([]byte(obj.Format()), ' ')
	result = strconv.AppendInt(result, int64(len(payload)), 10)
	result = append(result, '\x00')
	result = append(result, payload...)

	rawsha := sha1.Sum(result)
	sha := hex.EncodeToString(rawsha[:])

	if repo != nil {
		p, err := repo.RepoFile(true, "objects", string(sha[:2]), string(sha[2:]))
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

// If name is HEAD, it will just resolve .git/HEAD;
// If name is a full hash, this hash is returned unmodified.
// If name looks like a short hash, it will collect objects whose full hash begin with this short hash.
// At last, it will resolve tags and branches matching name.
func FindObject(repo *Repository, name, format string, follow bool) string {
	shas := resolveObject(repo, name)
	if len(shas) == 0 {
		return ""
	}
	if len(shas) > 1 {
		util.PanicErr(fmt.Errorf("Ambiguous reference %s: Candidates are:\n - %v", name, strings.Join(shas, "\b - ")))
		return ""
	}

	sha := shas[0]
	if format == "" {
		return sha
	}

	for {
		obj := ReadObject(repo, sha)
		if obj.Format() == format {
			return sha
		}

		if !follow {
			return ""
		}

		if obj.Format() == "tag" {
			sha = obj.(*TagObj).Object
		} else if obj.Format() == "commit" && format == "tree" {
			sha = obj.(*CommitObj).Tree
		} else {
			return ""
		}
	}
}

/*
	Resolve name to an object hash in repo.

This function is aware of:

  - the HEAD literal
  - short and long hashes
  - tags
  - branches
  - remote branches
*/
func resolveObject(repo *Repository, name string) []string {
	candidates := make([]string, 0)
	hashRegx := regexp.MustCompile("^[0-9A-Fa-f]{4,40}$")
	name = strings.TrimSpace(name)

	if name == "" {
		return nil
	}

	if name == "HEAD" {
		sha := GetRefSha(repo, "HEAD")
		if sha != "" {
			candidates = append(candidates, sha)
		}
		return candidates
	}

	if hashRegx.Match([]byte(name)) {
		name = strings.ToLower(name)
		prefix := name[:2]
		p, err := repo.repoDir(false, "objects", prefix)
		util.PanicErr(err)

		if p != "" {
			rem := name[2:]
			entries, err := os.ReadDir(p)
			util.PanicErr(err)
			for _, f := range entries {
				if strings.HasPrefix(f.Name(), rem) {
					candidates = append(candidates, prefix+f.Name())
				}
			}
		}
	}

	asTag := GetRefSha(repo, "refs/tags/"+name)
	if asTag != "" {
		candidates = append(candidates, asTag)
	}

	asBranch := GetRefSha(repo, "refs/heads/"+name)
	if asBranch != "" {
		candidates = append(candidates, asBranch)
	}

	return candidates
}
