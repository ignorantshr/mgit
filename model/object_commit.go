package model

import (
	"bytes"
	"strings"
)

// “Key-Value List with Message” for commit and tag files
type kvlm struct {
	kv       map[string]string
	Author   string
	Commiter string
	Tree     string
	Parent   string
	Gpgsign  string
	Message  string
}

func (k *kvlm) parse(raw []byte) {
	if k == nil {
		*k = kvlm{kv: make(map[string]string)}
	}
	if k.kv == nil {
		k.kv = map[string]string{}
	}

	for {
		spaceidx := bytes.IndexByte(raw, ' ')
		nlidx := bytes.IndexByte(raw, '\n')

		// A blank line means the remainder of the data is the message.
		if nlidx == 0 {
			k.kv["message"] = string(raw[nlidx+1:])
			return
		}

		key := string(raw[:spaceidx])
		end := spaceidx + 1
		for raw[end+1] != ' ' { // 值跨行时每行前面有一个空格
			end = bytes.IndexByte(raw[end:], '\n')
		}
		value := strings.ReplaceAll(string(raw[spaceidx+1:end]), "\n ", "\n")
		k.kv[key] += value // 防止覆盖旧值

		raw = raw[end+1:]
	}
}

func (k *kvlm) serialize() []byte {
	res := bytes.Buffer{}

	for key, value := range k.kv {
		if key == "message" {
			continue
		}

		res.WriteString(key)
		res.WriteByte(' ')
		res.WriteString(strings.ReplaceAll(value, "\n", "\n "))
		res.WriteByte('\n')
	}

	res.WriteString(k.kv["message"])
	return res.Bytes()
}

type CommitObj struct {
	fmt  string
	data []byte
	*kvlm
}

func NewCommitObj(data []byte) *CommitObj {
	return &CommitObj{"commit", data, &kvlm{}}
}

func (c *CommitObj) KV() map[string]string {
	return c.kv
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
