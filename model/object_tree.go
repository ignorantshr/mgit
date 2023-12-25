package model

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"
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
}

func parseTreeLeaf(raw []byte) (int, *treeLeaf) {
	leaf := &treeLeaf{}
	space := bytes.IndexByte(raw, ' ')
	if space != 6 && space != 5 {
		util.PanicErr(fmt.Errorf("invalid tree file"))
	}

	leaf.Mode = string(raw[:space])
	if space == 5 {
		leaf.Mode = " " + leaf.Mode
	}

	null := bytes.IndexByte(raw, '\x00')
	leaf.Path = string(raw[space+1 : null])

	leaf.Sha = hex.EncodeToString(raw[null+1 : null+21])
	return null + 21, leaf
}
