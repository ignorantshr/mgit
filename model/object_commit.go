package model

import (
	"bytes"
	"fmt"
	"strings"
)

type CommitObj struct {
	fmt string
	*kvlm
}

func NewCommitObj() *CommitObj {
	return &CommitObj{"commit", &kvlm{}}
}

func (c *CommitObj) KV() *kvlm {
	return c.kvlm
}

func (c *CommitObj) Format() string {
	return c.fmt
}

func (c *CommitObj) Serialize(_ *Repository) []byte {
	return c.kvlm.serialize()
}

func (c *CommitObj) Deserialize(data []byte) {
	c.kvlm.parse(data)
}

// “Key-Value List with Message” for commit and tag files
type kvlm struct {
	// common
	Message string

	// commit
	Author   string
	Commiter string
	Tree     string
	Parent   string
	Gpgsign  string

	// tag
	Object string
	Type   string
	Tag    string
	Tagger string
}

func (k *kvlm) parse(raw []byte) {
	if k == nil {
		*k = kvlm{}
	}

	for {
		spaceidx := bytes.IndexByte(raw, ' ')
		nlidx := bytes.IndexByte(raw, '\n')

		// A blank line means the remainder of the data is the message.
		if nlidx == 0 {
			k.Message = string(raw[nlidx+1:])
			return
		}

		key := string(raw[:spaceidx])
		end := spaceidx + 1
		for raw[end+1] != ' ' { // 值跨行时每行前面有一个空格
			end = bytes.IndexByte(raw[end:], '\n')
		}
		value := strings.ReplaceAll(string(raw[spaceidx+1:end]), "\n ", "\n")
		switch key {
		case "tree":
			k.Tree = value
		case "parent":
			k.Parent += value
		case "author":
			k.Author = value
		case "commiter":
			k.Commiter = value
		case "gpgsig":
			k.Gpgsign = value
		}

		raw = raw[end+1:]
	}
}

func (k *kvlm) serialize() []byte {
	res := bytes.Buffer{}

	write := func(k, v string) {
		if len(v) == 0 {
			return
		}
		res.WriteString(fmt.Sprintf("%s: ", k))
		res.WriteString(strings.ReplaceAll(v, "\n", "\n "))
		res.WriteByte('\n')
	}

	write("tree", k.Tree)
	write("author", k.Author)
	write("commiter", k.Commiter)
	write("gpgsig", k.Gpgsign)
	res.WriteByte('\n')
	res.WriteString(k.Message)
	return res.Bytes()
}
