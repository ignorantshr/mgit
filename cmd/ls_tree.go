package cmd

import (
	"fmt"
	"path"
	"strings"

	"github.com/ignorantshr/mgit/model"
	"github.com/ignorantshr/mgit/util"
	"github.com/spf13/cobra"
)

var _recur *bool

func init() {
	_recur = lsTreeCmd.Flags().BoolP("recursive", "r", false, "recurse into sub-trees")
	rootCmd.AddCommand(lsTreeCmd)
}

var lsTreeCmd = &cobra.Command{
	Use:   "ls-tree TREE",
	Short: "Pretty-print a tree object.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repo := model.FindRepo(".")
		lsTree(repo, args[0], "", *_recur)
	},
}

func lsTree(repo *model.Repository, ref, prefix string, recursive bool) {
	sha := model.FindObject(repo, ref, "tree", true)
	if sha == "" {
		return
	}
	obj := model.ReadObject(repo, sha).(*model.TreeObj)

	typ := ""
	for _, v := range obj.Items() {
		if len(v.Mode) == 5 {
			typ = v.Mode[:1]
		} else {
			typ = v.Mode[:2]
		}
		switch typ {
		case "04":
			typ = "tree"
		case "10": // file
			typ = "blob"
		case "12": // link
			typ = "blob"
		case "16":
			typ = "commit"
		default:
			util.PanicErr(fmt.Errorf("unacknowledged format: %s", typ))
		}

		if recursive && typ == "tree" {
			lsTree(repo, ref, path.Join(prefix, v.Path), recursive)
		} else { // leaf
			fmt.Printf("%v %v %v\t%v\n", strings.Repeat("0", 6-len(v.Mode))+v.Mode, typ, v.Sha, path.Join(prefix, v.Path))
		}
	}
}
