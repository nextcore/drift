// Package fetch provides HTTP download utilities and GitHub API interactions
// for fetching prebuilt Skia binaries.
package fetch

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// VerifyChecksum computes the SHA256 hash of the file and compares it to the expected value.
// Returns nil if the checksum matches, or an error with details if it doesn't.
func VerifyChecksum(filePath, expectedSHA256 string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for checksum: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("failed to read file for checksum: %w", err)
	}

	actual := hex.EncodeToString(h.Sum(nil))
	if actual != expectedSHA256 {
		return &ChecksumError{
			File:     filePath,
			Expected: expectedSHA256,
			Actual:   actual,
		}
	}

	return nil
}

// ChecksumError is returned when a file's checksum doesn't match the expected value.
type ChecksumError struct {
	File     string
	Expected string
	Actual   string
}

func (e *ChecksumError) Error() string {
	return fmt.Sprintf("checksum mismatch for %s\nExpected: %s\nActual:   %s", e.File, e.Expected, e.Actual)
}
