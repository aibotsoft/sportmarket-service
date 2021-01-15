package token

import (
	"time"
)

type Token struct {
	Session     string
	CreatedAt   time.Time
	LastCheckAt time.Time
}

func NewToken() Token {
	return Token{}
}

func (t *Token) IsEmpty() bool {
	return t.Session == ""
}
