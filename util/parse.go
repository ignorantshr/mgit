package util

import (
	"encoding/binary"
)

// func IntToBytes(num int, width int) []byte {
// 	buf := make([]byte, width)
// 	switch width {
// 	case 1:
// 		buf[0] = byte(num & 0xFF)
// 	case 2:
// 		binary.BigEndian.PutUint16(buf, uint16(num))
// 	case 4:
// 		binary.BigEndian.PutUint32(buf, uint32(num))
// 	case 8:
// 		binary.BigEndian.PutUint64(buf, uint64(num))
// 	default:
// 		PanicErr(errors.New("不支持的宽度"))
// 	}
// 	return buf
// }

// func BytesToInt(data []byte) int {
// 	switch len(data) {
// 	case 1:
// 		return int(data[0])
// 	case 2:
// 		return int(binary.BigEndian.Uint16(data))
// 	case 4:
// 		return int(binary.BigEndian.Uint32(data))
// 	case 8:
// 		return int(binary.BigEndian.Uint64(data))
// 	default:
// 		PanicErr(errors.New("不支持的宽度"))
// 		return 0
// 	}
// }

// func BytesToInt64(data []byte) int64 {
// 	switch len(data) {
// 	case 1:
// 		return int64(data[0])
// 	case 2:
// 		return int64(binary.BigEndian.Uint16(data))
// 	case 4:
// 		return int64(binary.BigEndian.Uint32(data))
// 	case 8:
// 		return int64(binary.BigEndian.Uint64(data))
// 	default:
// 		PanicErr(errors.New("不支持的宽度"))
// 		return 0
// 	}
// }

func IntToBytes(num int, width int) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(num))
	return buf[8-width:]
}

func BytesToInt(data []byte) int {
	buf := make([]byte, 8)
	copy(buf[8-len(data):], data)
	return int(binary.BigEndian.Uint64(buf))
}

func BytesToInt64(data []byte) int64 {
	buf := make([]byte, 8)
	copy(buf[8-len(data):], data)
	return int64(binary.BigEndian.Uint64(buf))
}
