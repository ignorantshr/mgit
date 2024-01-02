package cmd

import (
	"fmt"

	"github.com/ignorantshr/mgit/model"
	"github.com/spf13/cobra"
)

/* git check-ignore

检查路径是否与 git ignore 文件规则匹配
*/

func init() {
	rootCmd.AddCommand(checkIgnoreCmd)
}

var checkIgnoreCmd = &cobra.Command{
	Use:   "check-ignore <path ...>",
	Short: "Check path(s) against ignore rules.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repo := model.FindRepo(".")
		checkIgnore(repo, args)
	},
}

func checkIgnore(repo *model.Repository, paths []string) {
	rules := model.ReadGitignore(repo)
	for _, p := range paths {
		if model.CheckIgnore(p, rules) {
			fmt.Println(p)
		}
	}
}
