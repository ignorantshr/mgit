package model

import (
	"bytes"
	"os"
	"strings"

	"github.com/ignorantshr/mgit/util"
)

const (
	BranchDir = "refs/heads/"
)

func GetActiveBranch(repo *Repository) string {
	rf, err := repo.RepoFile(false, "HEAD")
	util.PanicErr(err)

	head, err := os.ReadFile(rf)
	util.PanicErr(err)

	if bytes.HasPrefix(head, []byte("ref: "+BranchDir)) {
		return strings.TrimSpace(string(head[16:])) // remove \n
	}
	return ""
}
