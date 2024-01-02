package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ignorantshr/mgit/model"
	"github.com/ignorantshr/mgit/util"
	"github.com/spf13/cobra"
)

var _branchAll bool
var _branchCopy bool
var _branchDelete bool

func init() {
	branchCmd.Flags().BoolVarP(&_branchAll, "all", "a", false, "list all local branches")
	branchCmd.Flags().BoolVarP(&_branchCopy, "copy", "c", false, "copy a new branch from a branch or a commit")
	branchCmd.Flags().BoolVarP(&_branchDelete, "delete", "d", false, "delete a branch")
	rootCmd.AddCommand(branchCmd)
}

var branchCmd = &cobra.Command{
	Use: "branch [-a] | " +
		"branch -c [<old-branch>|<commit>] <new-branch>",
	Short: "List and create and remove branches",
	Run: func(cmd *cobra.Command, args []string) {
		repo := model.FindRepo(".")
		if _branchCopy {
			oldName := "HEAD"
			newName := ""
			switch len(args) {
			case 1:
				newName = args[0]
			case 2:
				oldName = args[0]
				newName = args[1]
			default:
				fmt.Println("invalid args")
				os.Exit(1)
			}

			branchCopy(repo, oldName, newName)
		} else if _branchDelete {
			if len(args) < 1 {
				fmt.Println("invalid args")
				os.Exit(1)
			}
			branchDelete(repo, args[0])
		} else {
			branchList(repo, _branchAll)
		}
	},
}

func branchList(repo *model.Repository, all bool) {
	b := model.GetActiveBranch(repo)
	if b != "" {
		sha := model.GetRefSha(repo, filepath.Join(model.BranchDir, b))
		fmt.Printf("* %v %v\n", b, sha[:7])
	} else {
		fmt.Printf("HEAD detached at %v.\n", model.FindObject(repo, "HEAD", "", true))
	}

	if !all {
		return
	}

	filepath.WalkDir(filepath.Join(repo.GitDir(), model.BranchDir),
		func(path string, d fs.DirEntry, err error) error {
			if !d.IsDir() && d.Name() != b {
				sha := model.GetRefSha(repo, filepath.Join(model.BranchDir, d.Name()))
				fmt.Printf("  %v %v\n", d.Name(), sha[:7])
			}
			return nil
		})
}

func branchCopy(repo *model.Repository, oldName, newName string) {
	sha := model.FindObject(repo, oldName, "", true)
	if sha == "" {
		return
	}

	name := filepath.Join(repo.GitDir(), model.BranchDir, newName)
	if util.IsFileExist(name) {
		fmt.Printf("branch %v is already exist.\n", newName)
		return
	}
	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY, 0644)
	util.PanicErr(err)
	defer f.Close()

	f.WriteString(sha)
}

func branchDelete(repo *model.Repository, oldName string) {
	name := filepath.Join(repo.GitDir(), model.BranchDir, oldName)
	os.Remove(name)
}
