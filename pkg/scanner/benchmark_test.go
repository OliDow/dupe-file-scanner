package scanner

import (
	"dupe-file-checker/internal/testutil"
	"fmt"
	"testing"
)

func BenchmarkScan10Files(b *testing.B) {
	tmpDir := b.TempDir()

	for i := 0; i < 5; i++ {
		testutil.CreateTestFile(fmt.Sprintf("%s/dup_%d.txt", tmpDir, i), "Duplicate content")
	}
	for i := 0; i < 5; i++ {
		testutil.CreateTestFile(fmt.Sprintf("%s/unique_%d.txt", tmpDir, i), fmt.Sprintf("Unique %d", i))
	}

	s := New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Scan(tmpDir)
	}
}

func BenchmarkScan100Files(b *testing.B) {
	tmpDir := b.TempDir()

	for i := 0; i < 50; i++ {
		testutil.CreateTestFile(fmt.Sprintf("%s/dup_%d.txt", tmpDir, i), "Duplicate content")
	}
	for i := 0; i < 50; i++ {
		testutil.CreateTestFile(fmt.Sprintf("%s/unique_%d.txt", tmpDir, i), fmt.Sprintf("Unique %d", i))
	}

	s := New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Scan(tmpDir)
	}
}

func BenchmarkQuickHashOnly(b *testing.B) {
	tmpDir := b.TempDir()

	for i := 0; i < 10; i++ {
		testutil.CreateTestFile(fmt.Sprintf("%s/file_%d.txt", tmpDir, i), fmt.Sprintf("Unique content %d", i))
	}

	files, _ := Walk(tmpDir)
	s := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sizeGroups := s.groupBySize(files)
		s.processQuickHashes(sizeGroups)
	}
}
