package ssh

import (
	"crypto/sha256"
	"encoding/base64"

	gossh "golang.org/x/crypto/ssh"
)

func fingerprintSHA256(key gossh.PublicKey) string {
	sum := sha256.Sum256(key.Marshal())
	return "SHA256:" + base64.StdEncoding.EncodeToString(sum[:])
}

func keyTypeName(key gossh.PublicKey) string {
	if key == nil {
		return "unknown"
	}
	return key.Type()
}
