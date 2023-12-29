package model

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/ignorantshr/mgit/util"
)

type IgnoreRule struct {
	Rule     string
	Excluded bool
}

// 解析起始字符
func parseRuleUnit(raw string) *IgnoreRule {
	raw = strings.TrimSpace(raw)

	if raw == "" || raw[0] == '#' {
		return nil
	}
	if raw[0] == '!' {
		return &IgnoreRule{raw[1:], false}
	}
	if raw[0] == '\\' {
		return &IgnoreRule{raw[1:], true}
	}
	return &IgnoreRule{raw, true}
}

func parseGitignoreRules(lines []string) []*IgnoreRule {
	res := make([]*IgnoreRule, 0)

	for _, l := range lines {
		r := parseRuleUnit(l)
		if r != nil {
			res = append(res, r)
		}
	}

	return res
}

type GitIgnore struct {
	Absolute []*IgnoreRule            // 不在项目源码路径下的 .gitignore 文件，例如 ~/.config/git/ignore 或 .git/info/exclude
	Scoped   map[string][]*IgnoreRule // 存在于各个目录下的 .gitignore
}

func ReadGitignore(repo *Repository) *GitIgnore {
	res := &GitIgnore{[]*IgnoreRule{}, map[string][]*IgnoreRule{}}

	readRules := func(file string) {
		if util.IsFileExist(file) {
			content, err := os.ReadFile(file)
			util.PanicErr(err)

			lines := []string{}
			scanner := bufio.NewScanner(bytes.NewReader(content))
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
			res.Absolute = append(res.Absolute, parseGitignoreRules(lines)...)
		}
	}
	repoFile := path.Join(repo.gitdir, "info/exlcude")
	readRules(repoFile)

	// global conf
	confHome, ok := os.LookupEnv("XDG_CONFIG_HOME")
	if !ok {
		confHome = os.ExpandEnv("~/.config")
	}
	globalFile := path.Join(confHome, "git/ignore")
	readRules(globalFile)

	// .gitignore files in the index
	indexF := ReadIndex(repo)

	for _, v := range indexF.Entries {
		if v.Name == ".gitignore" || strings.HasSuffix(v.Name, "/.gitignore") {
			dir := path.Dir(v.Name)
			contents := ReadObject(repo, v.Sha).(*BlobObj)
			lines := strings.Split(string(contents.data), "\n")
			res.Scoped[dir] = parseGitignoreRules(lines)
		}
	}

	return res
}

func CheckIgnore(p string, rules *GitIgnore) bool {
	if path.IsAbs(p) {
		util.PanicErr(fmt.Errorf("requires path to be relative to the repository's root"))
	}

	res := checkIgnoreScoped(p, rules.Scoped)
	if res != nil {
		return *res
	}

	return checkIgnoreAbsolute(p, rules.Absolute)
}

// 检查在本工作树下的忽视规则
func checkIgnoreScoped(p string, rules map[string][]*IgnoreRule) *bool {
	parent := path.Dir(p)
	once := false // compatiable with .vscode/
	for {
		if set, ok := rules[parent]; ok {
			if res := checkIgnoreBase(p, set); res != nil {
				return res
			}
		}
		parent = path.Dir(parent)
		if parent == "." {
			if once {
				break
			}
			once = true
		}
	}
	return nil
}

// 检查在全局忽视规则
func checkIgnoreAbsolute(p string, rules []*IgnoreRule) bool {
	if res := checkIgnoreBase(p, rules); res != nil {
		return *res
	}
	return false
}

func checkIgnoreBase(p string, rules []*IgnoreRule) *bool {
	res := new(bool)
	for _, v := range rules {
		if ok, err := regexp.Match(v.Rule, []byte(p)); err != nil {
			continue
		} else if ok {
			*res = v.Excluded
			break
		}
	}
	return res
}
