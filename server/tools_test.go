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
		{"Test 1", "Hello, World!", "HelloWorld"},
		{"Test 2", "123 456 789", "123456789"},
		{"Test 3", "Special@#Characters!!", "SpecialCharacte"},
		{"Test 4", "Short", "Short"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeText(tt.input)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCreateDir(t *testing.T) {
	// Create a temporary directory
	dir := t.TempDir()
	err := CreateDir(dir)
	if err != nil {
		t.Errorf("CreateDir failed: %v", err)
	}
}

func TestSha1Hash(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{"Test 1", []byte("Hello, World!"), "0a0a9f2a6772942557ab5355d76af442f8f65e01"},
		{"Test 2", []byte("123 456 789"), "4a87373297b683dbcd9ce7de4d28edae6b514f92"},
		{"Test 3", []byte("xxxxxxx@xxxxxxxxxxxxx"), "b5e9d06e9b674f512a7391fc704c015d949715bc"},
		{"Test 4", []byte("Short"), "0fe7d82f25a3015040a206e54f9c1d3a9717c4c4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sha1Hash(tt.input)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestMergeMaps(t *testing.T) {
	m1 := map[string]string{"a": "1", "b": "2"}
	m2 := map[string]string{"c": "3", "d": "4"}
	expected := map[string]string{"a": "1", "b": "2", "c": "3", "d": "4"}

	result := MergeMaps(m1, m2)
	if len(result) != len(expected) {
		t.Errorf("got %v, want %v", result, expected)
	}
	for k, v := range expected {
		if result[k] != v {
			t.Errorf("got %v, want %v", result, expected)
		}
	}
}

func TestCopyImages(t *testing.T) {
	// Create a temporary directory
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create a file in the source directory
	file, err := os.Create(filepath.Join(srcDir, "test.txt"))
	if err != nil {
		t.Errorf("Create file failed: %v", err)
	}
	defer file.Close()

	// Copy the file
	_, err = CopyImages(srcDir, dstDir)
	if err != nil {
		t.Errorf("CopyImages failed: %v", err)
	}
}
