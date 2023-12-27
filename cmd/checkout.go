package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/ignorantshr/mgit/model"
	"github.com/ignorantshr/mgit/util"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(checkoutCmd)
}

var checkoutCmd = &cobra.Command{
	Use:                   "checkout <commit> <path>",
	Short:                 "Checkout a commit inside of a directory.",
	DisableFlagsInUseLine: true,
	Args:                  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		repo := model.FindRepo(".")
		obj := model.ReadObject(repo, model.FindObject(repo, args[0], "", true))
		if obj == nil {
			return
		}
		if obj.Format() == "commit" {
			cobj := obj.(*model.CommitObj)
			obj = model.ReadObject(repo, cobj.KV().Tree)
		}
		p := args[1]
		if !util.IsDir(p) || !util.IsDirEmpty(p) {
			util.PanicErr(fmt.Errorf("%s not a valid path", p))
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
