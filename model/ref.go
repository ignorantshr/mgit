package model

import (
	"os"
	"path"
	"strings"

	"github.com/ignorantshr/mgit/util"
)

func GetRefSha(repo *Repository, ref string) string {
	p, err := repo.repoFile(false, ref)
	if err != nil {
		util.PanicErr(err)
	}

	if !util.IsFile(p) {
		return ""
	}

	raw, err := os.ReadFile(p)
	if err != nil {
		util.PanicErr(err)
	}
	data := string(raw)
	if strings.HasPrefix(data, "ref: ") {
		return GetRefSha(repo, data[5:])
	}
	return data
}

func ListRef(repo *Repository, p string) map[string]any {
	if p == "" {
		var err error
		p, err = repo.repoDir(false, "refs")
		util.PanicErr(err)
	}

	entries, err := os.ReadDir(p)
	if err != nil {
		util.PanicErr(err)
	}

	res := make(map[string]any)
	for _, v := range entries {
		can := path.Join(p, v.Name())
		if util.IsDir(can) {
			res[v.Name()] = ListRef(repo, can)
		} else {
			res[v.Name()] = GetRefSha(repo, can)
		}
	}
	return res
}
