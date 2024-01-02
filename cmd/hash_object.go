package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/ignorantshr/mgit/model"
	"github.com/ignorantshr/mgit/util"
	"github.com/spf13/cobra"
)

/* git hash-object

计算目标 hash 值，存储到 .git
*/

var (
	writeFlag bool
	typeFlag  string
)

func init() {
	hashObjectCmd.Flags().BoolVarP(&writeFlag, "write", "w", false, "Actually write the object into the database")
	hashObjectCmd.Flags().StringVarP(&typeFlag, "type", "t", "", "Specify the type")
	hashObjectCmd.MarkFlagRequired("type")
	rootCmd.AddCommand(hashObjectCmd)
}

var hashObjectCmd = &cobra.Command{
	Use:   "hash-object {-t blob|tree|commit|tag} <file>",
	Short: "Compute object ID and optionally creates a blob from a file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var repo *model.Repository
		if writeFlag {
			repo = model.FindRepo(".")
		}
		sha := hashObject(args[0], typeFlag, repo)
		fmt.Println(sha)
	},
}

func hashObject(file string, format string, repo *model.Repository) string {
	raw, err := os.ReadFile(file)
	util.PanicErr(err)

	var obj model.Object
	switch format {
	case "blob":
		obj = model.NewBlobObj()
	case "tree":
		obj = model.NewBlobObj()
	case "commit":
		obj = model.NewCommitObj()
	case "tag":
		obj = model.NewTagObj()
	default:
		util.PanicErr(errors.New("unsupported format " + format))
	}

	obj.Deserialize(raw)
	return model.WriteObject(repo, obj)
}
