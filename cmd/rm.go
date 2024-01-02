package cmd

import (
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
	Use:   "rm <path ...>",
	Short: "Remove files from the working tree and the index.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repo := model.FindRepo(".")
		rm(repo, args)
	},
}

func rm(repo *model.Repository, paths []string) {
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

	for _, e := range index.Entries {
		fullPath := path.Join(repo.Worktree(), e.Name)
		if _, ok := abspaths[fullPath]; ok {
			delete(abspaths, fullPath)
		} else {
			keptEntries = append(keptEntries, e)
		}
	}

	index.Entries = keptEntries
	model.WriteIndex(repo, index)
}
