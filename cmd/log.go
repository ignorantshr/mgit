package cmd

import (
	"fmt"
	"strings"

	"github.com/ignorantshr/mgit/model"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(logCmd)
}

var logCmd = &cobra.Command{
	Use:   "log [commit]",
	Short: "Display history of a given commit",
	Run: func(cmd *cobra.Command, args []string) {
		repo := model.FindRepo(".")
		var sha string
		if len(args) >= 1 {
			sha = args[0]
		} else {
			sha = "HEAED"
		}
		sha = model.FindObject(repo, sha, "", true)
		logPrint(repo, sha)
	},
}

func logPrint(repo *model.Repository, sha string) {
	commit := model.ReadObject(repo, sha).(*model.CommitObj)
	kv := commit.KV()
	msg := kv["message"]

	fmt.Println("commit", sha)
	fmt.Println("Author:", kv["author"])
	fmt.Println("Date:", kv["date"])
	fmt.Println()
	fmt.Println("\t", msg)
	fmt.Println()

	parent, ok := kv["parent"]
	if !ok {
		return
	}

	ps := strings.Split(parent, ",")
	for _, p := range ps {
		logPrint(repo, p)
	}
}
