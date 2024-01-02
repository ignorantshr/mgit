package cmd

import (
	"fmt"

	"github.com/ignorantshr/mgit/model"
	"github.com/spf13/cobra"
)

func init() {
	revParseCmd.Flags().StringVarP(&revParseType, "type", "t", "", "Specify the expected type")
	rootCmd.AddCommand(revParseCmd)
}

var revParseType string

var revParseCmd = &cobra.Command{
	Use:                   "rev-parse <name>",
	Short:                 "Parse revision (or other objects) identifiers",
	DisableFlagsInUseLine: true,
	Args:                  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repo := model.FindRepo(args[0])
		revParse(repo, args[0], revParseType)
	},
}

func revParse(repo *model.Repository, name, format string) {
	fmt.Println(model.FindObject(repo, name, format, true))
}
