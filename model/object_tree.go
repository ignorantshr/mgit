package model

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/ignorantshr/mgit/util"
)

type TreeObj struct {
	fmt   string
	items []*treeLeaf
}

func NewTreeObj() *TreeObj {
	return &TreeObj{fmt: "tree"}
}

func (t *TreeObj) Items() []*treeLeaf {
	return t.items
}

func (t *TreeObj) Format() string {
	return t.fmt
}

func (t *TreeObj) Serialize(repo *Repository) []byte {
	return serializeTree(t.items)
}

func (t *TreeObj) Deserialize(data []byte) {
	t.items = parseTree(data)
}

// [mode] space [path] 0x00 [sha-1]
type treeLeaf struct {
	Mode string
	Path string
	Sha  string
}

func parseTree(raw []byte) []*treeLeaf {
	res := make([]*treeLeaf, 0)
	for len(raw) > 0 {
		pos, lf := parseTreeLeaf(raw)
		res = append(res, lf)
		raw = raw[pos:]
	}
	return res
}

func serializeTree(items []*treeLeaf) []byte {
	sort.Slice(items, func(i, j int) bool {
		// keep dir sort after file
		ifile := strings.HasPrefix(items[i].Path, "10")
		jfile := strings.HasPrefix(items[j].Path, "10")

		if ifile && jfile || !ifile && !jfile {
			return items[i].Path < items[j].Path
		}

		return ifile
	})

	res := bytes.Buffer{}
	for _, item := range items {
		res.WriteString(item.Mode)
		res.WriteByte(' ')
		res.WriteString(item.Path)
		res.WriteByte('\x00')
		sha, err := hex.DecodeString(item.Sha)
		util.PanicErr(err)
		res.Write(sha)
	}

	return res.Bytes()
	// return res
}

func parseTreeLeaf(raw []byte) (int, *treeLeaf) {
	leaf := &treeLeaf{}
	space := bytes.IndexByte(raw, ' ')
	if space != 6 {
		util.PanicErr(fmt.Errorf("invalid tree file"))
	}

	leaf.Mode = strconv.Itoa(util.BytesToInt(raw[:space]))
	if len(leaf.Mode) == 5 {
		leaf.Mode = "0" + leaf.Mode
	}

	null := space + 1 + bytes.IndexByte(raw[space+1:], '\x00')
	leaf.Path = string(raw[space+1 : null])

	leaf.Sha = hex.EncodeToString(raw[null+1 : null+21])
	return null + 21, leaf
}

// 树扁平化
func Tree2Map(repo *Repository, ref string, prefix string) map[string]string {
	res := make(map[string]string)
	sha := FindObject(repo, ref, "tree", true)
	if sha == "" {
		return res
	}
	tree := ReadObject(repo, sha).(*TreeObj)

	for _, leaf := range tree.items {
		fullPath := path.Join(prefix, leaf.Path)
		if strings.HasPrefix(leaf.Mode, "04") {
			subTree := Tree2Map(repo, leaf.Sha, fullPath)
			for k, v := range subTree {
				res[k] = v
			}
		} else {
			res[fullPath] = leaf.Sha
		}
	}

	return res
}
