package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSanitizeText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple alphanumeric text",
			input:    "HelloWorld123",
			expected: "HelloWorld123",
		},
		{
			name:     "Text with spaces",
			input:    "Hello World",
			expected: "HelloWorld",
		},
		{
			name:     "Text with special characters",
			input:    "Hello@World!#123",
			expected: "HelloWorld123",
		},
		{
			name:     "Text longer than 15 characters",
			input:    "ThisIsAVeryLongTextThatNeedsToBeTruncated",
			expected: "ThisIsAVeryLong",
		},
		{
			name:     "Text with only special characters",
			input:    "!@#$%^&*()",
			expected: "",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Text with numbers and symbols",
			input:    "ABC-123-XYZ",
			expected: "ABC123XYZ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeText(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeText(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSha1Hash(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "Empty data",
			input:    []byte{},
			expected: "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		},
		{
			name:     "Simple text",
			input:    []byte("hello world"),
			expected: "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed",
		},
		{
			name:     "Numeric data",
			input:    []byte("12345"),
			expected: "8cb2237d0679ca88db6464eac60da96345513964",
		},
		{
			name:     "JSON data",
			input:    []byte(`{"key":"value"}`),
			expected: "228458095a9502070fc113d99504226a6ff90a9a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sha1Hash(tt.input)
			if result != tt.expected {
				t.Errorf("Sha1Hash() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestMergeMaps(t *testing.T) {
	tests := []struct {
		name     string
		map1     map[string]string
		map2     map[string]string
		expected map[string]string
	}{
		{
			name: "Two non-overlapping maps",
			map1: map[string]string{"a": "1", "b": "2"},
			map2: map[string]string{"c": "3", "d": "4"},
			expected: map[string]string{
				"a": "1",
				"b": "2",
				"c": "3",
				"d": "4",
			},
		},
		{
			name: "Overlapping maps - second map wins",
			map1: map[string]string{"a": "1", "b": "2"},
			map2: map[string]string{"b": "3", "c": "4"},
			expected: map[string]string{
				"a": "1",
				"b": "3",
				"c": "4",
			},
		},
		{
			name:     "Empty maps",
			map1:     map[string]string{},
			map2:     map[string]string{},
			expected: map[string]string{},
		},
		{
			name:     "First map empty",
			map1:     map[string]string{},
			map2:     map[string]string{"a": "1"},
			expected: map[string]string{"a": "1"},
		},
		{
			name:     "Second map empty",
			map1:     map[string]string{"a": "1"},
			map2:     map[string]string{},
			expected: map[string]string{"a": "1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeMaps(tt.map1, tt.map2)

			if len(result) != len(tt.expected) {
				t.Errorf("MergeMaps() length = %d, want %d", len(result), len(tt.expected))
			}

			for key, expectedValue := range tt.expected {
				if resultValue, ok := result[key]; !ok {
					t.Errorf("MergeMaps() missing key %q", key)
				} else if resultValue != expectedValue {
					t.Errorf("MergeMaps()[%q] = %q, want %q", key, resultValue, expectedValue)
				}
			}
		})
	}
}

func TestCreateDir(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		dir       string
		shouldErr bool
	}{
		{
			name:      "Create single directory",
			dir:       filepath.Join(tmpDir, "test1"),
			shouldErr: false,
		},
		{
			name:      "Create nested directories",
			dir:       filepath.Join(tmpDir, "test2", "nested", "deep"),
			shouldErr: false,
		},
		{
			name:      "Create already existing directory",
			dir:       filepath.Join(tmpDir, "test3"),
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For the "already existing" test, create the directory first
			if tt.name == "Create already existing directory" {
				if err := os.MkdirAll(tt.dir, 0755); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			err := CreateDir(tt.dir)
			if (err != nil) != tt.shouldErr {
				t.Errorf("CreateDir() error = %v, shouldErr %v", err, tt.shouldErr)
				return
			}

			// Verify directory exists
			if !tt.shouldErr {
				if _, err := os.Stat(tt.dir); os.IsNotExist(err) {
					t.Errorf("CreateDir() did not create directory %q", tt.dir)
				}
			}
		})
	}
}

func TestCopyImages(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	dstDir := filepath.Join(tmpDir, "dst")

	// Create source directory
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Create test files in source directory
	testFiles := map[string][]byte{
		"image1.png": []byte("fake image 1 content"),
		"image2.jpg": []byte("fake image 2 content"),
		"logo.svg":   []byte("fake logo content"),
	}

	for filename, content := range testFiles {
		filePath := filepath.Join(srcDir, filename)
		if err := os.WriteFile(filePath, content, 0644); err != nil {
			t.Fatalf("Failed to create test file %q: %v", filename, err)
		}
	}

	// Test copying images
	t.Run("Copy images successfully", func(t *testing.T) {
		manifest, err := CopyImages(srcDir, dstDir)
		if err != nil {
			t.Fatalf("CopyImages() error = %v", err)
		}

		// Check that all files are in the manifest
		if len(manifest) != len(testFiles) {
			t.Errorf("CopyImages() manifest length = %d, want %d", len(manifest), len(testFiles))
		}

		// Verify each file was copied and manifest entry is correct
		for filename, originalContent := range testFiles {
			// Check file exists in destination
			dstFilePath := filepath.Join(dstDir, filename)
			copiedContent, err := os.ReadFile(dstFilePath)
			if err != nil {
				t.Errorf("Copied file %q not found: %v", filename, err)
				continue
			}

			// Check content matches
			if string(copiedContent) != string(originalContent) {
				t.Errorf("Copied file %q content does not match original", filename)
			}

			// Check manifest hash
			expectedHash := Sha1Hash(originalContent)
			if manifestHash, ok := manifest[filename]; !ok {
				t.Errorf("Manifest missing entry for %q", filename)
			} else if manifestHash != expectedHash {
				t.Errorf("Manifest hash for %q = %q, want %q", filename, manifestHash, expectedHash)
			}
		}
	})

	t.Run("Copy from non-existent directory", func(t *testing.T) {
		_, err := CopyImages(filepath.Join(tmpDir, "nonexistent"), filepath.Join(tmpDir, "dst2"))
		if err == nil {
			t.Error("CopyImages() expected error for non-existent source directory, got nil")
		}
	})
}

func TestReadRequestBody(t *testing.T) {
	// This function is already tested implicitly through integration tests
	// since it simply wraps io.ReadAll. We'll add a basic test for completeness.
	t.Skip("ReadRequestBody is a simple wrapper - tested in integration tests")
}
