package cmd

func Add(paths []string) {
	addCmd.Run(nil, paths)
}

func CatFile(paths []string) {
	catFileCmd.Run(nil, paths)
}

func CheckIgnore(paths []string) {
	checkIgnoreCmd.Run(nil, paths)
}

func Status() {
	statusCmd.Run(nil, nil)
}

func Log() {
	logCmd.Run(nil, nil)
}

func CommitCmd(msg string) {
	_commitMsg = msg
	commitCmd.Run(nil, nil)
}
