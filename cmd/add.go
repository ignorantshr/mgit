package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/ignorantshr/mgit/model"
	"github.com/ignorantshr/mgit/util"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:   "add <PAHT ...>",
	Short: "Add files contents to the index.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repo := model.FindRepo(".")
		add(repo, args, true)
	},
}

func add(repo *model.Repository, paths []string, realDelete bool) {
	rm(repo, paths, false)

	worktree := repo.Worktree() + string(filepath.Separator)

	cleanPaths := [][2]string{} // (absolute, relative_to_worktree)
	for _, p := range paths {
		p, _ = filepath.Abs(p)
		if strings.HasPrefix(p, worktree) && util.IsFile(p) {
			relp, _ := filepath.Rel(worktree, p)
			cleanPaths = append(cleanPaths, [2]string{p, relp})
		}
	}

	index := model.ReadIndex(repo)

	for _, p := range cleanPaths {
		sha := hashObject(p[0], "blob", repo)

		stat, err := os.Stat(p[0])
		util.PanicErr(err)
		fstat := stat.Sys().(*syscall.Stat_t)

		index.Entries = append(index.Entries, &model.IndexEntry{
			Ctime:          model.TimePair{S: fstat.Ctimespec.Sec, NS: fstat.Ctimespec.Nsec},
			Mtime:          model.TimePair{S: fstat.Mtimespec.Sec, NS: fstat.Mtimespec.Nsec},
			Device:         int64(fstat.Dev),
			Inode:          int64(fstat.Ino),
			ModeType:       0b1000,
			ModePerms:      0o644,
			Uid:            int(fstat.Uid),
			Gid:            int(fstat.Gid),
			Fsize:          stat.Size(),
			Sha:            sha,
			FlagAssumValid: false,
			FlagStage:      int64(fstat.Flags & 0b0011000000000000),
			Name:           p[1],
		})
	}

	model.WriteIndex(repo, index)
}
