package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/ignorantshr/mgit/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

/* git init 命令实现

创建 .git 文件目录结构，初始化 git 项目
*/

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a git directory",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := CreateRepository(args[0])
		if err != nil {
			panic(err)
		}
	},
}

const GitDir = ".mgit"

type repository struct {
	worktree string
	gitdir   string
	conf     *viper.Viper
}

func CreateRepository(p string) (*repository, error) {
	repo, err := newRepository(p, true)
	if err != nil {
		return nil, err
	}

	stat, err := os.Stat(repo.worktree)
	if err != nil {
		return nil, err
	}

	if !stat.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", repo.worktree)
	}

	if dir, err := os.ReadDir(repo.gitdir); err != nil {
		return nil, err
	} else if len(dir) != 0 {
		return nil, fmt.Errorf("%s is not empty", repo.gitdir)
	}

	if _, err := repo.repoDir(true, "branches"); err != nil {
		return nil, err
	}
	if _, err := repo.repoDir(true, "objects"); err != nil {
		return nil, err
	}
	if _, err := repo.repoDir(true, "refs", "tags"); err != nil {
		return nil, err
	}
	if _, err := repo.repoDir(true, "refs", "heads"); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(repo.repoPath("HEAD"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	f.WriteString("ref: refs/heads/master\n")

	f2, err := os.OpenFile(repo.repoPath("conf"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer f2.Close()
	f2.WriteString("[core]\n")
	f2.WriteString("repositoryformatversion = 0\n")
	f2.WriteString("filemode = false\n")
	f2.WriteString("bare = false\n")

	return repo, nil
}

func newRepository(p string, force bool) (*repository, error) {
	r := &repository{}
	r.worktree = p
	r.gitdir = path.Join(p, "GitDir")

	fileinfo, err := os.Stat(r.gitdir)

	if !force && (os.IsNotExist(err) || !fileinfo.IsDir()) {
		return nil, errors.New("not a Git repository")
	}

	cf, err := r.repoFile(false, "conf")
	if err != nil {
		return nil, err
	}
	_, err = os.Stat(cf)
	if err != nil {
		if os.IsNotExist(err) {
			return r, nil
		}
		return nil, err
	}

	r.conf = viper.New()
	r.conf.SetConfigName("conf")
	r.conf.SetConfigType("ini")
	r.conf.AddConfigPath(r.gitdir)
	if err := r.conf.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok && !force {
			return nil, fmt.Errorf("%s/conf file not found", GitDir)
		} else {
			return nil, err
		}
	}

	if !force {
		if vers := r.conf.GetInt("core.repositoryformatversion"); vers != 0 {
			return nil, fmt.Errorf("unsupported repositoryformatversion %d", vers)
		}
	}

	return r, nil
}

// 组装成 .git/** 文件字符串
func (r *repository) repoPath(paths ...string) string {
	return path.Join(r.gitdir, path.Join(paths...))
}

// 组装 .git/** 文件字符串，如果父目录缺失则创建目录结构
func (r *repository) repoFile(mkdir bool, paths ...string) (string, error) {
	if _, err := r.repoDir(mkdir, paths[:len(paths)-1]...); err != nil {
		return "", err
	} else {
		return r.repoPath(paths...), nil
	}
}

// 创建目录结构
func (r *repository) repoDir(mkdir bool, paths ...string) (string, error) {
	p := r.repoPath(paths...)
	return p, os.MkdirAll(p, 0755)
}

func FindRepo(p string) (*repository, error) {
	if util.IsDir(path.Join(p, GitDir)) {
		return newRepository(p, false)
	}

	pp := filepath.Dir(p)
	if pp == "/" {
		return nil, fmt.Errorf("no git directory")
	}

	return FindRepo(pp)
}
