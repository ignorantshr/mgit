package cmd

import (
	"testing"
)

func Test_checkIgnore(t *testing.T) {
	checkIgnoreCmd.Run(nil, []string{"go.mod", "mgit", ".vscode", ".mgit", "cmd/cat_file.go"})
}
