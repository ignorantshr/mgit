package model

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"os"
	"path"
	"sort"
	"strconv"

	"github.com/ignorantshr/mgit/util"
)

type TimePair struct {
	S  int64
	NS int64
}

type IndexEntry struct {
	Ctime          TimePair
	Mtime          TimePair
	Device         int64 // the ID of device containing this file
	Inode          int64 // inode
	ModeType       int   // The object type, either b1000 (regular), b1010 (symlink), b1110 (gitlink)
	ModePerms      int   // The object permissions, an integer
	Uid            int
	Gid            int
	Fsize          int64
	Sha            string
	FlagAssumValid bool
	FlagStage      int64
	Name           string // full path
}

type Index struct {
	Version int
	Entries []*IndexEntry
}

func NewIndex(ver int, entries []*IndexEntry) *Index {
	if ver < 2 {
		ver = 2
	}
	i := &Index{Version: ver}
	if len(entries) == 0 {
		i.Entries = make([]*IndexEntry, 0)
	} else {
		i.Entries = entries
	}
	return i
}

func ReadIndex(repo *Repository) *Index {
	indexFile, err := repo.RepoFile(false, "index")
	util.PanicErr(err)

	index := NewIndex(2, nil)
	// new repository have no index
	if !util.IsFile(indexFile) {
		return index
	}

	raw, err := os.ReadFile(indexFile)
	util.PanicErr(err)

	header := raw[:12]
	signature := header[:4]
	if string(signature) != "DIRC" {
		util.PanicErr(fmt.Errorf("index file header error: %v", signature))
	}
	version := util.BytesToInt(header[4:8])
	if version != 2 {
		util.PanicErr(fmt.Errorf("index file version error: %v", version))
	}
	count := util.BytesToInt(header[8:])

	content := raw[12:]
	idx := int64(0)
	for i := 0; i < count; i++ {
		ctime_s := util.BytesToInt64(content[idx : idx+4])
		ctime_ns := util.BytesToInt64(content[idx+4 : idx+8])
		mtime_s := util.BytesToInt64(content[idx+8 : idx+12])
		mtime_ns := util.BytesToInt64(content[idx+12 : idx+16])
		device := util.BytesToInt64(content[idx+16 : idx+20])
		ino := util.BytesToInt64(content[idx+20 : idx+24])
		// unused := util.BytesToInt(content[idx+24:idx+26]) // assert
		mode := util.BytesToInt(content[idx+26 : idx+28])
		modeType := mode >> 12
		modePerms := mode & 0b0000000111111111

		uid := util.BytesToInt(content[idx+28 : idx+32])
		gid := util.BytesToInt(content[idx+32 : idx+36])
		fsize := util.BytesToInt64(content[idx+36 : idx+40])
		sha := fmt.Sprintf("%40x", new(big.Int).SetBytes(content[idx+40:idx+60])) // 将字节序列表示的整数转换为一个 40 字节长度的十六进制字符串
		flags := util.BytesToInt64(content[idx+60 : idx+62])
		flagAssumValid := (flags & 0b1000000000000000) != 0
		// flagExtend := (flags & 0b0100000000000000) != 0 // assert
		flagStage := flags & 0b0011000000000000
		nameLength := flags & 0b0000111111111111 // 12bit 存储，最大 0xFFF，可能会溢出，所以继续向后寻找直到 0x00
		idx += 62
		var rawName []byte
		if nameLength < 0xFFF {
			if content[idx+nameLength] != 0x00 {
				util.PanicErr(fmt.Errorf("name length parse fail"))
			}
			rawName = content[idx : idx+nameLength]
			idx += nameLength + 1
		} else {
			nullIdx := bytes.IndexByte(content[idx+0xFFF:], '\x00')
			rawName = content[idx : idx+0xFFF+int64(nullIdx)]
			idx = int64(nullIdx) + 1
		}
		name := string(rawName)
		idx = 8 * int64(math.Ceil(float64(idx)/8))
		index.Entries = append(index.Entries, &IndexEntry{
			Ctime:          TimePair{int64(ctime_s), ctime_ns},
			Mtime:          TimePair{mtime_s, mtime_ns},
			Device:         device,
			Inode:          ino,
			ModeType:       int(modeType),
			ModePerms:      int(modePerms),
			Uid:            int(uid),
			Gid:            int(gid),
			Fsize:          fsize,
			Sha:            sha,
			FlagAssumValid: flagAssumValid,
			FlagStage:      flagStage,
			Name:           name,
		})
	}

	return index
}

func WriteIndex(repo *Repository, index *Index) {
	p, err := repo.RepoFile(false, "index")
	util.PanicErr(err)
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	util.PanicErr(err)
	defer f.Close()

	// 字节宽度，int 值
	wrinteger := func(width, value int) {
		f.Write(util.IntToBytes(value, width))
	}

	// HEADER
	f.WriteString("DIRC")
	wrinteger(4, index.Version)
	wrinteger(4, len(index.Entries))

	// ENTRIES

	idx := 0
	for _, e := range index.Entries {
		wrinteger(4, int(e.Ctime.S))
		wrinteger(4, int(e.Ctime.NS))
		wrinteger(4, int(e.Mtime.S))
		wrinteger(4, int(e.Mtime.NS))
		wrinteger(4, int(e.Device))
		wrinteger(4, int(e.Inode))

		wrinteger(4, int(e.ModeType<<12|e.ModePerms))
		wrinteger(4, int(e.Uid))
		wrinteger(4, int(e.Gid))
		wrinteger(4, int(e.Fsize))

		sha, _ := hex.DecodeString(e.Sha)
		f.Write(sha)

		flagAssumValid := 0
		if e.FlagAssumValid {
			flagAssumValid = 0x1 << 15
		}
		nameLen := len(e.Name)
		if nameLen >= 0xFFF {
			nameLen = 0xFFF
		}
		wrinteger(2, flagAssumValid|int(e.FlagStage)|nameLen)
		f.WriteString(e.Name)

		// 0x00
		wrinteger(1, 0)

		idx += 62 + nameLen + 1

		if idx%8 != 0 {
			pad := 8 - idx%8
			buf := make([]byte, pad)
			f.Write(buf)
			idx += pad
		}
	}
}

func Index2Tree(repo *Repository, index *Index) string {
	contents := make(map[string][]any)
	sortpaths := make([]string, 0)

	for _, e := range index.Entries {
		dir := path.Dir(e.Name)

		key := dir
		for key != "." {
			if _, ok := contents[key]; !ok {
				contents[key] = []any{}
			}
			key = path.Dir(key)
		}

		contents[dir] = append(contents[dir], e)
		sortpaths = append(sortpaths, dir)
	}

	sort.Slice(sortpaths, func(i, j int) bool {
		return len(sortpaths[i]) > len(sortpaths[j])
	})

	sha := ""
	for _, p := range sortpaths {
		tree := NewTreeObj()

		for _, e := range contents[p] {
			// An entry can be a normal GitIndexEntry read from the index,
			// or a tree we've created.
			var leaf *treeLeaf
			switch v := e.(type) {
			case *IndexEntry:
				mode, _ := strconv.Atoi(fmt.Sprintf("%02o%04o", v.ModeType, v.ModePerms))
				leafMode := string(util.IntToBytes(mode, 6))
				leaf = &treeLeaf{Mode: leafMode, Path: path.Base(v.Name), Sha: v.Sha}
			case [2]string:
				mode, _ := strconv.Atoi(fmt.Sprintf("%02o%04o", 4, 0))
				leafMode := string(util.IntToBytes(mode, 6))
				leaf = &treeLeaf{Mode: leafMode, Path: path.Base(v[0]), Sha: v[1]}
			}
			tree.items = append(tree.items, leaf)
		}

		sha = WriteObject(repo, tree)
		parent := path.Dir(p)
		base := path.Base(p)
		contents[parent] = append(contents[parent], [2]string{base, sha})
	}

	return sha
}
