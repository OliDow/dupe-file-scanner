package scanner

import (
	"dupe-file-checker/internal/testutil"
	"testing"
)

func TestScanNoDuplicates(t *testing.T) {
	tmpDir := t.TempDir()

	testutil.CreateTestFile(tmpDir+"/file1.txt", "Content 1")
	testutil.CreateTestFile(tmpDir+"/file2.txt", "Different content")
	testutil.CreateTestFile(tmpDir+"/file3.txt", "Yet another unique content")

	s := New()
	duplicates, err := s.Scan(tmpDir, false)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(duplicates) != 0 {
		t.Errorf("Expected 0 duplicate groups, got %d", len(duplicates))
	}
}

func TestScanWithDuplicates(t *testing.T) {
	tmpDir := t.TempDir()

	content := "Duplicate content for testing"
	testutil.CreateTestFile(tmpDir+"/file1.txt", content)
	testutil.CreateTestFile(tmpDir+"/file2.txt", content)
	testutil.CreateTestFile(tmpDir+"/nested/file3.txt", content)

	s := New()
	duplicates, err := s.Scan(tmpDir, false)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(duplicates) != 1 {
		t.Fatalf("Expected 1 duplicate group, got %d", len(duplicates))
	}

	if len(duplicates[0].Files) != 3 {
		t.Errorf("Expected 3 duplicate files, got %d", len(duplicates[0].Files))
	}
}

func TestScanEmptyFiles(t *testing.T) {
	tmpDir := t.TempDir()

	testutil.CreateTestFile(tmpDir+"/empty1.txt", "")
	testutil.CreateTestFile(tmpDir+"/empty2.txt", "")

	s := New()
	duplicates, err := s.Scan(tmpDir, false)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(duplicates) != 1 {
		t.Fatalf("Expected 1 duplicate group for empty files, got %d", len(duplicates))
	}

	if len(duplicates[0].Files) != 2 {
		t.Errorf("Expected 2 empty duplicate files, got %d", len(duplicates[0].Files))
	}
}

func TestScanMultipleGroups(t *testing.T) {
	tmpDir := t.TempDir()

	testutil.CreateTestFile(tmpDir+"/group1_a.txt", "Group 1 content")
	testutil.CreateTestFile(tmpDir+"/group1_b.txt", "Group 1 content")

	testutil.CreateTestFile(tmpDir+"/group2_a.txt", "Group 2 content")
	testutil.CreateTestFile(tmpDir+"/group2_b.txt", "Group 2 content")
	testutil.CreateTestFile(tmpDir+"/group2_c.txt", "Group 2 content")

	testutil.CreateTestFile(tmpDir+"/unique.txt", "Unique content")

	s := New()
	duplicates, err := s.Scan(tmpDir, false)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(duplicates) != 2 {
		t.Fatalf("Expected 2 duplicate groups, got %d", len(duplicates))
	}
}

func TestScanSameSizeDifferentContent(t *testing.T) {
	tmpDir := t.TempDir()

	testutil.CreateTestFile(tmpDir+"/file1.txt", "Content AAA")
	testutil.CreateTestFile(tmpDir+"/file2.txt", "Content BBB")

	s := New()
	duplicates, err := s.Scan(tmpDir, false)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(duplicates) != 0 {
		t.Errorf("Expected 0 duplicates for same-size different content, got %d", len(duplicates))
	}
}

func TestScanWithTestFixtures(t *testing.T) {
	tmpDir := t.TempDir()

	if err := testutil.SetupTestFixtures(tmpDir); err != nil {
		t.Fatalf("Failed to setup test fixtures: %v", err)
	}

	s := New()
	duplicates, err := s.Scan(tmpDir, false)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(duplicates) < 2 {
		t.Errorf("Expected at least 2 duplicate groups (identical files + empty files), got %d", len(duplicates))
	}

	foundIdenticalGroup := false
	foundEmptyGroup := false

	for _, group := range duplicates {
		if len(group.Files) == 3 {
			foundIdenticalGroup = true
		}
		if len(group.Files) == 2 {
			foundEmptyGroup = true
		}
	}

	if !foundIdenticalGroup {
		t.Error("Expected to find group of 3 identical files")
	}

	if !foundEmptyGroup {
		t.Error("Expected to find group of 2 empty files")
	}
}
