package model

// 存储用户文件数据，不需要特殊处理
type BlobObj struct {
	fmt  string
	data []byte
}

func NewBlobObj() Object {
	return &BlobObj{fmt: "blob"}
}

func (b *BlobObj) Format() string {
	return b.fmt
}

func (b *BlobObj) Serialize(_ *Repository) []byte {
	return b.data
}

func (b *BlobObj) Deserialize(data []byte) {
	b.data = data
}
