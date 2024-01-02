package cmd

import (
	"os"
	"path"
	"path/filepath"
	"slices"
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
	Use:   "add <path ...>",
	Short: "Add files contents to the index.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repo := model.FindRepo(".")
		add(repo, args, true)
	},
}

// 删除新增或修改的旧条目，然后重写 index 文件
func add(repo *model.Repository, paths []string, realDelete bool) {
	pathSet := expandPaths(repo, nil, paths)
	addedPath := []string{}

	for k := range pathSet {
		addedPath = append(addedPath, k)
	}

	rm(repo, addedPath)

	worktree := repo.Worktree() + string(filepath.Separator)

	cleanPaths := [][2]string{} // (absolute, relative_to_worktree)
	for p := range pathSet {
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

func expandPaths(repo *model.Repository, rules *model.GitIgnore, paths []string) map[string]struct{} {
	if rules == nil {
		rules = model.ReadGitignore(repo)
	}

	dir := []string{}
	res := make(map[string]struct{})
	for _, p := range paths {
		if !model.CheckIgnore(p, rules) {
			abso, _ := filepath.Abs(p)
			if !util.IsDir(abso) {
				res[p] = struct{}{}
			} else if !util.IsDirEmpty(abso) {
				dir = append(dir, p)
			}
		}
	}

	dir = slices.Compact[[]string](dir)
	child := []string{}
	for _, d := range dir {
		entries, err := os.ReadDir(d)
		util.PanicErr(err)
		for _, e := range entries {
			child = append(child, path.Join(d, e.Name()))
		}
	}
	if len(child) > 0 {
		sub := expandPaths(repo, rules, child)
		for p := range sub {
			res[p] = struct{}{}
		}
	}

	return res
}
