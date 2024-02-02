package main

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SanitizeText sanitizes the text to be used as a filename.
// It removes all non-alphanumeric characters, replaces spaces with underscores, and truncates to 15 characters.
func SanitizeText(text string) string {
	// Remove all non-alphanumeric characters.
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	sanitized := reg.ReplaceAllString(text, "")

	// Remove spaces.
	sanitized = strings.ReplaceAll(sanitized, " ", "_")

	// Truncate to 15 characters.
	if len(sanitized) > 15 {
		sanitized = sanitized[:15]
	}

	return sanitized
}

// CreateDir creates a directory if it doesn't exist
func CreateDir(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return nil
}

// Sha1Hash returns the SHA1 hash of the given data as a hex string
func Sha1Hash(data []byte) string {
	hash := sha1.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

// MergeMaps merges two maps
func MergeMaps(m1 map[string]string, m2 map[string]string) map[string]string {
	merged := make(map[string]string)
	for k, v := range m1 {
		merged[k] = v
	}
	for key, value := range m2 {
		merged[key] = value
	}
	return merged
}

// CopyImages copies all images from the source directory to the destination directory and returns a manifest of the images
func CopyImages(srcDir, dstDir string) (map[string]string, error) {
	// Ensure destination directory exists.
	if err := CreateDir(dstDir); err != nil {
		return nil, err
	}

	// Get list of files in source directory.
	files, err := os.ReadDir(srcDir)
	if err != nil {
		return nil, err
	}

	// Manifest for images
	manifest := make(map[string]string)

	// Iterate over each file.
	for _, file := range files {
		// Skip directories.
		if file.IsDir() {
			continue
		}

		// Open source file.
		srcFile, err := os.Open(filepath.Join(srcDir, file.Name()))
		if err != nil {
			return nil, err
		}
		defer srcFile.Close()

		// Create destination file.
		newFilePath := filepath.Join(dstDir, file.Name())
		dstFile, err := os.Create(newFilePath)
		if err != nil {
			return nil, err
		}
		defer dstFile.Close()

		// Copy content from source file to destination file.
		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return nil, err
		}

		// Read the content of the destination file.
		dstContent, err := os.ReadFile(newFilePath)
		if err != nil {
			return nil, err
		}

		manifest[file.Name()] = Sha1Hash(dstContent)
	}

	return manifest, nil
}
