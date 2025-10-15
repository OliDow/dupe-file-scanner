package testutil

import (
	"os"
	"path/filepath"
)

func CreateTestFile(path string, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func SetupTestFixtures(baseDir string) error {
	fixtures := map[string]string{
		"identical/file1.txt":            "This is identical content for testing duplicates.\n",
		"identical/file1_copy.txt":       "This is identical content for testing duplicates.\n",
		"identical/nested/file1_dup.txt": "This is identical content for testing duplicates.\n",
		"similar/file2.txt":              "This is different content with same length---\n",
		"similar/file2_diff.txt":         "This is another text with the same length!!!\n",
		"edge_cases/empty.txt":           "",
		"edge_cases/empty_dup.txt":       "",
		"edge_cases/small.txt":           "x",
		"edge_cases/small_dup.txt":       "x",
		"unique/file3.txt":               "Unique content that has no duplicates anywhere.\n",
	}

	for relPath, content := range fixtures {
		fullPath := filepath.Join(baseDir, relPath)
		if err := CreateTestFile(fullPath, content); err != nil {
			return err
		}
	}

	return nil
}

func CleanupTestFixtures(baseDir string) error {
	return os.RemoveAll(baseDir)
}
