package cmd

import (
	"fmt"
	"math"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/ignorantshr/mgit/model"
	"github.com/ignorantshr/mgit/util"
	"github.com/spf13/cobra"
)

/* git status

通过将 index 文件和 HEAD 做对比、将 index 文件和 文件系统 做对比 实现。
*/

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the working tree status.",
	Run: func(cmd *cobra.Command, args []string) {
		repo := model.FindRepo(".")
		status(repo)
	},
}

func status(repo *model.Repository) {
	index := model.ReadIndex(repo)

	statusBranch(repo)
	statusHeadIndex(repo, index)
	statusHeadWorktree(repo, index)
}

func statusBranch(repo *model.Repository) {
	branch := model.GetActiveBranch(repo)
	if branch != "" {
		fmt.Printf("On branch %v.\n", branch)
	} else {
		fmt.Printf("HEAD detached at %v.\n", model.FindObject(repo, "HEAD", "", true))
	}
}

// Finding changes between HEAD and index
// 将 index 文件和 HEAD 做对比，对比出将要提交的更改类型
func statusHeadIndex(repo *model.Repository, index *model.Index) {
	if len(index.Entries) != 0 {
		fmt.Println("Changes to be committed:")
	}

	head := model.Tree2Map(repo, "HEAD", "")
	for _, v := range index.Entries {
		if sha, ok := head[v.Name]; ok {
			if sha != v.Sha {
				fmt.Printf("\tmodified: %s\n", v.Name)
			}
			delete(head, v.Name)
		} else {
			fmt.Printf("\tadded: %s\n", v.Name)
		}
	}

	for k := range head {
		fmt.Printf("\tdeleted: %s", k)
	}
	fmt.Println()
}

// 将 index 文件和 文件系统 做对比，找出没有处于 stage 的更改
func statusHeadWorktree(repo *model.Repository, index *model.Index) {
	ignore := model.ReadGitignore(repo)

	allFiles := walkFilesystem(repo)

	// traverse the index, and compare real files with the cached versions.
	if len(index.Entries) != 0 {
		fmt.Println("Changes not staged for commit:")
	}

	modified := []string{}
	deleted := []string{}
	for _, v := range index.Entries {
		fullPath := path.Join(repo.Worktree(), v.Name)

		if !util.IsFileExist(fullPath) {
			deleted = append(deleted, v.Name)
		} else {
			stat, _ := os.Stat(fullPath)
			ctimeNS := v.Ctime.S*int64(math.Pow10(9)) + v.Ctime.NS
			mtimeNS := v.Mtime.S*int64(math.Pow10(9)) + v.Mtime.NS
			fstat := stat.Sys().(*syscall.Stat_t)

			// 将操作系统特定的时间戳转换为 Go 中的时间类型
			ctime := time.Unix(int64(fstat.Ctimespec.Sec), int64(fstat.Ctimespec.Nsec))
			if ctimeNS != ctime.UnixNano() || mtimeNS != stat.ModTime().UnixNano() {
				newSha := hashObject(fullPath, "blob", nil)
				if newSha != v.Sha {
					modified = append(modified, v.Name)
				}
			}
		}

		delete(allFiles, v.Name)
	}

	sort.Strings(modified)
	for _, name := range modified {
		fmt.Printf("\tmodified: %v\n", name)
	}

	sort.Strings(deleted)
	for _, name := range deleted {
		fmt.Printf("\tdeleted: %v\n", name)
	}

	untracked := []string{}
	if len(allFiles) != 0 {
		fmt.Println()
		fmt.Println("Untracked files:")

		for f := range allFiles {
			if !model.CheckIgnore(f, ignore) {
				untracked = append(untracked, f)
			}
		}
	}

	sort.Strings(untracked)
	for _, name := range untracked {
		fmt.Printf("\t%v\n", name)
	}
}

// 记录仓库下所有的文件
func walkFilesystem(repo *model.Repository) map[string]struct{} {
	allFiles := make(map[string]struct{}, 0)

	var cur string
	queue := make([]string, 0)
	queue = append(queue, repo.Worktree())
	for len(queue) != 0 {
		cur = queue[0]
		subFiles, err := os.ReadDir(cur)
		util.PanicErr(err)

		for _, v := range subFiles {
			if strings.HasPrefix(cur, repo.GitDir()) || v.Name() == ".git" || v.Name() == model.GitDir {
				continue
			}

			fullPath := path.Join(cur, v.Name())
			if !v.IsDir() {
				relPath, err := filepath.Rel(repo.Worktree(), fullPath) // 获取从前者到后者的相对路径
				util.PanicErr(err)
				allFiles[relPath] = struct{}{}
			} else {
				queue = append(queue, fullPath)
			}
		}
		queue = queue[1:]
	}

	return allFiles
}
