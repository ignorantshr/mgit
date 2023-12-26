package model

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"os"
	"strconv"

	"github.com/ignorantshr/mgit/util"
)

type timePair struct {
	S  int64
	NS int64
}

type IndexEntry struct {
	Ctime          timePair
	Mtime          timePair
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
	indexFile, err := repo.repoFile(false, "index")
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
	version, _ := strconv.Atoi(string(header[4:8]))
	if version != 2 {
		util.PanicErr(fmt.Errorf("index file version error: %v", version))
	}
	count, _ := strconv.Atoi(string(header[8:])) // how much entries

	content := raw[12:]
	idx := int64(0)
	for i := 0; i < count; i++ {
		ctime_s, _ := strconv.ParseInt(string(content[idx:idx+4]), 10, 64)
		ctime_ns, _ := strconv.ParseInt(string(content[idx+4:idx+8]), 10, 64)
		mtime_s, _ := strconv.ParseInt(string(content[idx+8:idx+12]), 10, 64)
		mtime_ns, _ := strconv.ParseInt(string(content[idx+12:idx+16]), 10, 64)
		device, _ := strconv.ParseInt(string(content[idx+16:idx+20]), 10, 64)
		ino, _ := strconv.ParseInt(string(content[idx+20:idx+24]), 10, 64)
		// unused, _ := strconv.ParseInt(string(content[idx+24:idx+26]), 10, 64) // assert
		mode, _ := strconv.ParseInt(string(content[idx+26:idx+28]), 10, 64)
		modeType := mode >> 12
		modePerms := modeType & 0b0000000111111111

		uid, _ := strconv.ParseInt(string(content[idx+28:idx+32]), 10, 64)
		gid, _ := strconv.ParseInt(string(content[idx+32:idx+36]), 10, 64)
		fsize, _ := strconv.ParseInt(string(content[idx+36:idx+40]), 10, 64)
		sha := fmt.Sprintf("%40x", new(big.Int).SetBytes(content[idx+40:idx+60])) // 将字节序列表示的整数转换为一个 40 字节长度的十六进制字符串
		flags, _ := strconv.ParseInt(string(content[idx+60:idx+62]), 10, 64)
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
			Ctime:          timePair{ctime_s, ctime_ns},
			Mtime:          timePair{mtime_s, mtime_ns},
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
