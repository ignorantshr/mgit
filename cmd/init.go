package cmd

import (
	"github.com/ignorantshr/mgit/model"
	"github.com/spf13/cobra"
)

/* git init 命令实现

创建 .git 文件目录结构，初始化 git 项目
*/

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init <path>",
	Short: "Initialize a git directory",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_, err := model.CreateRepository(args[0])
		if err != nil {
			panic(err)
		}
	},
}
