package hasher

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComputeQuickHash(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "This is test content for quick hash"

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	info, _ := os.Stat(testFile)
	qh, err := ComputeQuickHash(testFile, info.Size(), info.ModTime().Unix())
	if err != nil {
		t.Fatalf("ComputeQuickHash failed: %v", err)
	}

	if qh.Size != info.Size() {
		t.Errorf("Expected size %d, got %d", info.Size(), qh.Size)
	}

	if qh.HeadHash == 0 {
		t.Error("Expected non-zero head hash")
	}
}

func TestComputeQuickHashIdentical(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	content := "Identical content"

	os.WriteFile(file1, []byte(content), 0644)
	os.WriteFile(file2, []byte(content), 0644)

	info1, _ := os.Stat(file1)
	info2, _ := os.Stat(file2)

	qh1, _ := ComputeQuickHash(file1, info1.Size(), info1.ModTime().Unix())
	qh2, _ := ComputeQuickHash(file2, info2.Size(), info2.ModTime().Unix())

	if qh1.HeadHash != qh2.HeadHash {
		t.Error("Expected identical files to have same head hash")
	}
}

func TestComputeFullHash(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "Content for full hash testing"

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	hash, err := ComputeFullHash(testFile)
	if err != nil {
		t.Fatalf("ComputeFullHash failed: %v", err)
	}

	if hash == 0 {
		t.Error("Expected non-zero hash")
	}
}

func TestComputeFullHashIdentical(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	content := "Identical content for hash comparison"

	os.WriteFile(file1, []byte(content), 0644)
	os.WriteFile(file2, []byte(content), 0644)

	hash1, _ := ComputeFullHash(file1)
	hash2, _ := ComputeFullHash(file2)

	if hash1 != hash2 {
		t.Errorf("Expected identical files to have same hash, got %d and %d", hash1, hash2)
	}
}

func TestComputeFullHashDifferent(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	os.WriteFile(file1, []byte("Content A"), 0644)
	os.WriteFile(file2, []byte("Content B"), 0644)

	hash1, _ := ComputeFullHash(file1)
	hash2, _ := ComputeFullHash(file2)

	if hash1 == hash2 {
		t.Error("Expected different files to have different hashes")
	}
}
