package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/ignorantshr/mgit/util"
	"github.com/spf13/cobra"
)

/* git hash-object 命令实现

计算目标 hash 值，存储到 .git
*/

var (
	writeFlag bool
	typeFlag  string
)

func init() {
	hashObjectCmd.Flags().BoolVarP(&writeFlag, "write", "w", false, "Actually write the object into the database")
	hashObjectCmd.Flags().StringVarP(&typeFlag, "type", "t", "blob", "Specify the type")
	rootCmd.AddCommand(hashObjectCmd)
}

var hashObjectCmd = &cobra.Command{
	Use:   "hash-object [-w] [-t TYPE] FILE",
	Short: "Compute object ID and optionally creates a blob from a file",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var repo *repository
		if writeFlag {
			dir, _ := os.Getwd()
			repo = FindRepo(dir)
		}
		sha := hashObject(args[0], typeFlag, repo)
		fmt.Println(sha)
	},
}

func hashObject(file string, format string, repo *repository) string {
	raw, err := os.ReadFile(file)
	util.PanicErr(err)

	var obj object
	switch format {
	case "blob":
		obj = newBlob(raw)
	default:
		util.PanicErr(errors.New("unsupported format " + format))
	}

	return writeObject(obj, repo)
}
