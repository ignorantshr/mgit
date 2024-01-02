package cmd

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/ignorantshr/mgit/model"
	"github.com/ignorantshr/mgit/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

/*
commit 实际上就是把 index 转换成 commit 对象

1. 把 index 转换成 tree 对象
2. 生成并存储相应的 commit 对象
3. 更新 HEAD 分支指向新的 commit（分支就是指向一个 commit 的引用）
*/

var _commitMsg string

func init() {
	commitCmd.Flags().StringVarP(&_commitMsg, "message", "m", "nothing", "Message to associate with this commit.")
	commitCmd.MarkFlagRequired("message")
	rootCmd.AddCommand(commitCmd)
}

var commitCmd = &cobra.Command{
	Use:   "commit -m <message>",
	Short: "Record changes to the repository.",
	Run: func(cmd *cobra.Command, args []string) {
		repo := model.FindRepo(".")
		commit(repo, _commitMsg)
	},
}

func commit(repo *model.Repository, msg string) {
	index := model.ReadIndex(repo)
	treesha := model.Index2Tree(repo, index)

	com := model.CreateCommit(repo, treesha, model.FindObject(repo, "HEAD", "", true), readGitAuthor(), msg, time.Now())

	ab := model.GetActiveBranch(repo)
	if ab != "" {
		p, err := repo.RepoFile(false, path.Join("refs/heads", ab))
		util.PanicErr(err)
		f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		util.PanicErr(err)
		defer f.Close()
		f.WriteString(model.WriteObject(repo, com) + "\n")
	} else {
		p, err := repo.RepoFile(false, "HEAD")
		util.PanicErr(err)
		f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		util.PanicErr(err)
		defer f.Close()
		f.WriteString("\n")
	}
}

func readGitAuthor() string {
	xdgConfigHome, ok := os.LookupEnv("XDG_CONFIG_HOME")
	userhome, _ := os.UserHomeDir()
	if !ok {
		xdgConfigHome = path.Join(userhome, ".conf")
	}

	vip := viper.New()
	p1 := path.Join(xdgConfigHome, "git/config")
	if util.IsFile(p1) {
		vip.SetConfigFile(p1)
	}
	p2 := path.Join(userhome, ".gitconfig")
	if util.IsFile(p2) {
		vip.SetConfigFile(p2)
	}
	vip.SetConfigType("ini")
	vip.ReadInConfig()
	user := vip.GetStringMap("user")
	name := user["name"]
	email := user["email"]
	return fmt.Sprintf("%s <%s>", name, email)
}
