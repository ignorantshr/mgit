package model

type TagObj struct {
	CommitObj
}

func NewTagObj() *TagObj {
	return &TagObj{CommitObj{"tag", &kvlm{Type: "commit"}}}
}
