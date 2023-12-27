package model

import (
	"bytes"
	"os"
	"strings"

	"github.com/ignorantshr/mgit/util"
)

func GetActiveBranch(repo *Repository) string {
	rf, err := repo.repoFile(false, "HEAD")
	util.PanicErr(err)

	head, err := os.ReadFile(rf)
	util.PanicErr(err)

	if bytes.HasPrefix(head, []byte("ref: refs/heads/")) {
		return strings.TrimSpace(string(head[16:])) // remove \n
	}
	return ""
}
