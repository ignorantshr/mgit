package cmd

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ignorantshr/mgit/model"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(rmCmd)
}

var rmCmd = &cobra.Command{
	Use:   "rm <PAHT ...>",
	Short: "Remove files from the working tree and the index.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repo := model.FindRepo(".")
		rm(repo, args, true)
	},
}

func rm(repo *model.Repository, paths []string, realDelete bool) {
	index := model.ReadIndex(repo)

	worktree := repo.Worktree() + string(filepath.Separator)
	abspaths := map[string]struct{}{}

	for _, p := range paths {
		p, _ = filepath.Abs(p)
		if strings.HasPrefix(p, worktree) {
			abspaths[p] = struct{}{}
		}
	}

	keptEntries := []*model.IndexEntry{}
	remove := []string{}

	for _, e := range index.Entries {
		fullPath := path.Join(repo.Worktree(), e.Name)
		if _, ok := abspaths[fullPath]; ok {
			remove = append(remove, fullPath)
			delete(abspaths, fullPath)
		} else {
			keptEntries = append(keptEntries, e)
		}
	}

	if realDelete {
		for _, p := range remove {
			os.RemoveAll(p)
		}
	}

	index.Entries = keptEntries
	model.WriteIndex(repo, index)
}
