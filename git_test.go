package main

import (
	"log/slog"
	"path"
	"regexp"
	"testing"

	"github.com/ignorantshr/mgit/cmd"
)

func TestAdd(t *testing.T) {
	cmd.Add([]string{"go.mod"})
}

func TestCheckIgnore(t *testing.T) {
	t.Log(regexp.MatchString(".vscode/", ".vscode/"))
	p := ".vscode/"
	c := path.Clean(p)
	parent := path.Dir(p)
	slog.Info("ignore", "pa", parent, "cp", c)
	cmd.CheckIgnore([]string{"mgit", "cmd/", "aaa", ".vscode/"})
}

func TestStatus(t *testing.T) {
	cmd.Status()
}

func TestLog(t *testing.T) {
	cmd.Log()
}

func TestCatFile(t *testing.T) {
	cmd.CatFile([]string{"tree", "09f289cc8000455c17bd04d0dbd974e382fd7bae"})
	// cmd.CatFile([]string{"blob", "1caadda4945ca8fa04e298fbeb3746d151a6a9fb"})
}

func TestCommit(t *testing.T) {
	msg := "ttt"
	cmd.CommitCmd(msg)
}
