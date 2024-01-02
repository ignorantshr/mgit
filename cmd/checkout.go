package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/ignorantshr/mgit/model"
	"github.com/ignorantshr/mgit/util"
	"github.com/spf13/cobra"
)

/* git checkout

因为 mgit 没有完整的测试过，所以为了防止对 worktree 造成不可恢复的动作，我们使用新的文件夹来承载历史版本
*/

func init() {
	rootCmd.AddCommand(checkoutCmd)
}

var checkoutCmd = &cobra.Command{
	Use:                   "checkout {<commit>|<branch>} <path>",
	Short:                 "Checkout a commit inside of a directory.",
	DisableFlagsInUseLine: true,
	Args:                  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		repo := model.FindRepo(".")
		p := args[1]
		if util.IsFile(p) || (util.IsDir(p) && !util.IsDirEmpty(p)) {
			util.PanicErr(fmt.Errorf("%s not a valid path", p))
		}

		sha := model.FindObject(repo, args[0], "", true)
		obj := model.ReadObject(repo, sha)
		if obj == nil {
			return
		}
		if obj.Format() == "commit" {
			cobj := obj.(*model.CommitObj)
			sha = cobj.KV().Tree
			obj = model.ReadObject(repo, sha)
		}

		os.MkdirAll(p, 0755)
		checkoutTree(repo, p, obj.(*model.TreeObj))
		buildGitDir(repo, args[0], p, sha)
	},
}

func checkoutTree(repo *model.Repository, destPath string, tree *model.TreeObj) {
	for _, v := range tree.Items() {
		obj := model.ReadObject(repo, v.Sha)
		dest := path.Join(destPath, v.Path)

		if obj.Format() == "tree" {
			err := os.Mkdir(dest, 0755)
			checkoutTree(repo, dest, obj.(*model.TreeObj))
			util.PanicErr(err)
		} else if obj.Format() == "blob" {
			err := os.WriteFile(dest, obj.(*model.BlobObj).Serialize(repo), 0644)
			util.PanicErr(err)
		}
	}
}

func buildGitDir(repo *model.Repository, src, destPath, ref string) {
	newGitDir := filepath.Join(destPath, model.GitDir)
	err := util.CopyDir(repo.GitDir(), newGitDir)
	util.PanicErr(err)
	os.Chdir(destPath)

	paths := []string{}
	entries := model.Tree2Map(repo, ref, "")
	for p := range entries {
		paths = append(paths, p)
	}
	repo.SetWorktree(filepath.Join(repo.Worktree(), destPath))
	repo.SetGitDir(filepath.Join(repo.Worktree(), model.GitDir))
	os.Remove(filepath.Join(repo.GitDir(), "index"))
	add(repo, paths)

	head, err := os.OpenFile(filepath.Join(repo.GitDir(), "HEAD"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	util.PanicErr(err)
	if model.HashRegx.MatchString(src) {
		head.WriteString(src)
	} else {
		head.WriteString("ref: " + model.BranchDir + src)
	}
}
