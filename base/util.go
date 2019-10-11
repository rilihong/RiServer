package base

import (
	"crypto/rand"
	"fmt"
)

type SByte []byte

func RandToken() string {
	b := make([]byte, 10)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}