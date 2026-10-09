package update

import (
	"testing"
)

const testPublicKey = "untrusted comment: minisign public key\nRWQf6LRCGA9i53mlYecO4IzT51TGPpvWucNSCh1CBM0QTaLn73Y7GFO3\n"

const testSigSmall = "untrusted comment: signature from minisign secret key\n" +
	"RWQf6LRCGA9i59SLOFxz6NxvASXDJeRtuZykwQepbDEGt87ig1BNpWaVWuNrm73YiIiJbq71Wi+dP9eKL8OC351vwIasSSbXxwA=\n" +
	"trusted comment: timestamp:1635442742\tfile:test\n" +
	"0YteLgV960ia80vnA/fHbvkyjl/IoP/HNOCaZfrF0CdhAlp7ok+Tpkya+VpWPX5C/Is3q8a/kEDSY7fBmmgJCg==\n"

func TestVerifySignedFile_ok(t *testing.T) {
	t.Parallel()
	payload := []byte("test")
	if err := VerifySignedFile(payload, []byte(testSigSmall), testPublicKey); err != nil {
		t.Fatalf("VerifySignedFile: %v", err)
	}
}

func TestVerifySignedFile_tampered(t *testing.T) {
	t.Parallel()
	payload := []byte("tampered")
	if err := VerifySignedFile(payload, []byte(testSigSmall), testPublicKey); err == nil {
		t.Fatal("expected error for tampered payload")
	}
}

func TestVerifyChecksumLine_ok(t *testing.T) {
	t.Parallel()
	artifact := []byte("hello")
	// sha256("hello") from spec
	line := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824  wisp_linux_amd64.tar.gz"
	if err := VerifyChecksumLine(line, artifact); err != nil {
		t.Fatalf("VerifyChecksumLine: %v", err)
	}
}

func TestVerifyChecksumLine_mismatch(t *testing.T) {
	t.Parallel()
	line := "0000000000000000000000000000000000000000000000000000000000000000  wisp_linux_amd64.tar.gz"
	if err := VerifyChecksumLine(line, []byte("hello")); err == nil {
		t.Fatal("expected checksum mismatch")
	}
}
