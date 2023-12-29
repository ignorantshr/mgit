package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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
			sha = "HEAD"
		}
		sha = model.FindObject(repo, sha, "", true)
		logPrint(repo, sha)
	},
}

func logPrint(repo *model.Repository, sha string) {
	if sha == "" {
		return
	}
	commit := model.ReadObject(repo, sha).(*model.CommitObj)
	kv := commit.KV()
	msg := kv.Message
	author := strings.Split(kv.Author, " ")
	ts, _ := strconv.Atoi(author[2])

	fmt.Println("commit", sha)
	fmt.Println("Author:", author[0], author[1])
	fmt.Println("Date:  ", time.Unix(int64(ts), 0))
	fmt.Println()
	fmt.Println("    " + msg)
	fmt.Println()

	parent := kv.Parent
	if len(parent) == 0 {
		return
	}

	ps := strings.Split(parent, "\n")
	for _, p := range ps {
		logPrint(repo, p)
	}
}
