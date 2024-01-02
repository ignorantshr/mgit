package cmd

import (
	"fmt"

	"github.com/ignorantshr/mgit/model"
	"github.com/spf13/cobra"
)

/* git show-ref

列出 git 所有的引用对象
*/

func init() {
	rootCmd.AddCommand(showRefCmd)
}

var showRefCmd = &cobra.Command{
	Use:   "show-ref",
	Short: "List references.",
	Run: func(cmd *cobra.Command, args []string) {
		repo := model.FindRepo(".")
		refs := model.ListRef(repo, "")
		showRef(repo, refs, true, "")
	},
}

func showRef(repo *model.Repository, refs map[string]any, withHash bool, prefix string) {
	p := prefix
	if prefix != "" {
		p += "/"
	}
	for k, v := range refs {
		switch v := v.(type) {
		case string:
			if withHash {
				v += " "
			} else {
				v = ""
			}
			fmt.Printf("%s%s%s\n", v, p, k)
		default:
			showRef(repo, v.(map[string]any), withHash, p+k)
		}
	}
}
