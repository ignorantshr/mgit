package cmd

import (
	"fmt"
	"os/user"
	"strconv"
	"time"

	"github.com/ignorantshr/mgit/model"
	"github.com/spf13/cobra"
)

var _lsFilesVerbose bool

func init() {
	lsFilesCmd.Flags().BoolVarP(&_lsFilesVerbose, "verbose", "v", false, "Show everything")
	rootCmd.AddCommand(lsFilesCmd)
}

var lsFilesCmd = &cobra.Command{
	Use:   "ls-files",
	Short: "List all the stage files",
	Run: func(cmd *cobra.Command, args []string) {
		repo := model.FindRepo(".")
		lsFiles(repo, _lsFilesVerbose)
	},
}

var _modeType = map[int]string{
	0b1000: "regular",
	0b1010: "symlink",
	0b1110: "gitlink",
}

func lsFiles(repo *model.Repository, verbose bool) {
	index := model.ReadIndex(repo)
	if verbose {
		fmt.Printf("Index file v%d, %d entries\n", index.Version, len(index.Entries))
	}

	for _, v := range index.Entries {
		fmt.Println(v.Name)
		if verbose {
			fmt.Printf("\t%v with perms: %v\n", _modeType[v.ModeType], v.ModePerms)
			fmt.Printf("\ton blob: %v\n", v.Sha)
			fmt.Printf("\tcreated: %v %v.%v, modified: %v %v.%v\n", time.Unix(v.Ctime.S, v.Ctime.NS), v.Ctime.S, v.Ctime.NS, time.Unix(v.Mtime.S, v.Mtime.NS), v.Mtime.S, v.Mtime.NS)
			fmt.Printf("\tdevice: %v, inode: %v\n", v.Device, v.Inode)
			u, _ := user.LookupId(strconv.Itoa(v.Uid))
			g, _ := user.LookupGroupId(strconv.Itoa(v.Gid))
			fmt.Printf("\tuser: %v(%v), group: %v(%v)\n", u, v.Uid, g.Name, v.Gid)
			fmt.Printf("\tflags: stage=%v assume_valid=%v\n", v.FlagStage, v.FlagAssumValid)
		}
	}
}
