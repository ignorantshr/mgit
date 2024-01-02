package cmd

import (
	"fmt"
	"os"
	"path"

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
	Use:                   "checkout {commit|branch} <path>",
	Short:                 "Checkout a commit inside of a directory.",
	DisableFlagsInUseLine: true,
	Args:                  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		repo := model.FindRepo(".")
		p := args[1]
		if util.IsFile(p) || (util.IsDir(p) && !util.IsDirEmpty(p)) {
			util.PanicErr(fmt.Errorf("%s not a valid path", p))
		}

		obj := model.ReadObject(repo, model.FindObject(repo, args[0], "", true))
		if obj == nil {
			return
		}
		if obj.Format() == "commit" {
			cobj := obj.(*model.CommitObj)
			obj = model.ReadObject(repo, cobj.KV().Tree)
		}

		os.MkdirAll(p, 0755)
		checkoutTree(repo, p, obj.(*model.TreeObj))
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
