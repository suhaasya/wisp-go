package server

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
)

func randomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

func randomHex(n int) (string, error) {
	b, err := randomBytes(n)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func strconvItoa(n int) string {
	return strconv.Itoa(n)
}
