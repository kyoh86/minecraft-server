package mclink

import (
	"crypto/rand"
	"strings"
)

const alphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

func NewCode(n int) (string, error) {
	if n <= 0 {
		n = 8
	}
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	var sb strings.Builder
	sb.Grow(n)
	for _, v := range b {
		sb.WriteByte(alphabet[int(v)%len(alphabet)])
	}
	return sb.String(), nil
}
