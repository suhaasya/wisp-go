package update

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"strings"

	"golang.org/x/crypto/blake2b"
)

type minisignPublicKey struct {
	SignatureAlgorithm [2]byte
	KeyId              [8]byte
	PublicKey          [32]byte
}

type minisignSignature struct {
	UntrustedComment   string
	SignatureAlgorithm [2]byte
	KeyId              [8]byte
	Signature          [64]byte
	TrustedComment     string
	GlobalSignature    [64]byte
}

func decodeMinisignPublicKey(in string) (minisignPublicKey, error) {
	lines := strings.SplitN(in, "\n", 2)
	if len(lines) < 2 {
		return minisignPublicKey{}, errors.New("incomplete encoded public key")
	}
	var publicKey minisignPublicKey
	bin, err := base64.StdEncoding.DecodeString(lines[1])
	if err != nil || len(bin) != 42 {
		return publicKey, errors.New("invalid encoded public key")
	}
	copy(publicKey.SignatureAlgorithm[:], bin[0:2])
	copy(publicKey.KeyId[:], bin[2:10])
	copy(publicKey.PublicKey[:], bin[10:42])
	return publicKey, nil
}

func decodeMinisignSignature(in string) (minisignSignature, error) {
	var signature minisignSignature
	lines := strings.SplitN(in, "\n", 4)
	if len(lines) < 4 {
		return signature, errors.New("incomplete encoded signature")
	}
	signature.UntrustedComment = strings.TrimRight(lines[0], "\r")
	bin1, err := base64.StdEncoding.DecodeString(lines[1])
	if err != nil || len(bin1) != 74 {
		return signature, errors.New("invalid encoded signature")
	}
	signature.TrustedComment = strings.TrimRight(lines[2], "\r")
	bin2, err := base64.StdEncoding.DecodeString(lines[3])
	if err != nil || len(bin2) != 64 {
		return signature, errors.New("invalid encoded signature")
	}
	copy(signature.SignatureAlgorithm[:], bin1[0:2])
	copy(signature.KeyId[:], bin1[2:10])
	copy(signature.Signature[:], bin1[10:74])
	copy(signature.GlobalSignature[:], bin2)
	return signature, nil
}

func verifyMinisign(publicKey minisignPublicKey, bin []byte, signature minisignSignature) (bool, error) {
	if publicKey.SignatureAlgorithm != [2]byte{'E', 'd'} {
		return false, errors.New("incompatible signature algorithm")
	}
	prehashed := false
	if signature.SignatureAlgorithm[0] == 0x45 && signature.SignatureAlgorithm[1] == 0x64 {
		prehashed = false
	} else if signature.SignatureAlgorithm[0] == 0x45 && signature.SignatureAlgorithm[1] == 0x44 {
		prehashed = true
	} else {
		return false, errors.New("unsupported signature algorithm")
	}
	if publicKey.KeyId != signature.KeyId {
		return false, errors.New("incompatible key identifiers")
	}
	if !strings.HasPrefix(signature.TrustedComment, "trusted comment: ") {
		return false, errors.New("unexpected format for the trusted comment")
	}
	if prehashed {
		h, _ := blake2b.New512(nil)
		h.Write(bin)
		bin = h.Sum(nil)
	}
	if !ed25519.Verify(ed25519.PublicKey(publicKey.PublicKey[:]), bin, signature.Signature[:]) {
		return false, errors.New("invalid signature")
	}
	if !ed25519.Verify(
		ed25519.PublicKey(publicKey.PublicKey[:]),
		append(signature.Signature[:], []byte(signature.TrustedComment)[17:]...),
		signature.GlobalSignature[:],
	) {
		return false, errors.New("invalid global signature")
	}
	return true, nil
}
