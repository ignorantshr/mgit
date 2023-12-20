package cmd

import (
	"fmt"

	"github.com/ignorantshr/mgit/model"
	"github.com/spf13/cobra"
)

/* git cat-file 命令实现

查看目标内容
*/

func init() {
	rootCmd.AddCommand(catFileCmd)
}

var catFileCmd = &cobra.Command{
	Use:   "cat-file {blob} object",
	Short: "Provide content of repository objects",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("requires at least %d arg(s), only received %d", 2, len(args))
		}
		switch args[0] {
		case "blob":
		default:
			return fmt.Errorf("unsupported format %s", args[0])
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		repo := model.FindRepo(".")
		catFile(repo, args[0], args[1])
	},
}

func catFile(repo *model.Repository, format, objStr string) {
	object := model.ReadObject(repo, model.FindObject(repo, objStr, format, true))
	fmt.Println(object.Serialize(nil))
}
