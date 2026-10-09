package update

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// VerifyChecksumLine compares artifact to a checksums.txt line: "<hex>  <filename>".
func VerifyChecksumLine(line string, artifact []byte) error {
	line = strings.TrimSpace(line)
	if line == "" {
		return errors.New("update: empty checksum line")
	}
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return errors.New("update: malformed checksum line")
	}
	wantHex := strings.ToLower(parts[0])
	sum := sha256.Sum256(artifact)
	gotHex := hex.EncodeToString(sum[:])
	if gotHex != wantHex {
		return fmt.Errorf("update: checksum mismatch for %s", parts[1])
	}
	return nil
}

// VerifySignedFile checks a payload against a minisign signature using publicKeyPEM (.pub file text).
func VerifySignedFile(payload, sigFile []byte, publicKeyPEM string) error {
	pk, err := decodeMinisignPublicKey(string(publicKeyPEM))
	if err != nil {
		return fmt.Errorf("update: public key: %w", err)
	}
	sig, err := decodeMinisignSignature(string(sigFile))
	if err != nil {
		return fmt.Errorf("update: signature: %w", err)
	}
	ok, err := verifyMinisign(pk, payload, sig)
	if err != nil {
		return fmt.Errorf("update: verify: %w", err)
	}
	if !ok {
		return errors.New("update: signature rejected")
	}
	return nil
}

// VerifyReleaseArchive verifies archive bytes against checksums.txt and minisign signature on checksums.
func VerifyReleaseArchive(archive []byte, archiveName string, checksums []byte, checksumSig []byte, publicKeyPEM string) error {
	if err := VerifySignedFile(checksums, checksumSig, publicKeyPEM); err != nil {
		return err
	}
	line, err := findChecksumLine(string(checksums), archiveName)
	if err != nil {
		return err
	}
	return VerifyChecksumLine(line, archive)
}

func findChecksumLine(checksums, name string) (string, error) {
	for _, line := range strings.Split(checksums, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[1] == name {
			return line, nil
		}
	}
	return "", fmt.Errorf("update: no checksum line for %q", name)
}
