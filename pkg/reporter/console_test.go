package reporter

import (
	"dupe-file-checker/pkg/scanner"
	"path/filepath"
	"runtime"
	"testing"
)

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{2147483648, "2.0 GB"},
	}

	for _, test := range tests {
		result := formatSize(test.bytes)
		if result != test.expected {
			t.Errorf("formatSize(%d) = %s; want %s", test.bytes, result, test.expected)
		}
	}
}

func TestExtractDirectory(t *testing.T) {
	// Use platform-specific paths
	var tests []struct {
		filePath string
		expected string
	}

	if runtime.GOOS == "windows" {
		tests = []struct {
			filePath string
			expected string
		}{
			{"C:\\Users\\John\\Documents\\file.txt", "C:\\Users\\John\\Documents"},
			{"C:\\Users\\John\\Downloads\\file.pdf", "C:\\Users\\John\\Downloads"},
			{"./file.txt", "(current directory)"},
			{"file.txt", "(current directory)"},
			{"C:\\file.txt", "C:\\"},
		}
	} else {
		tests = []struct {
			filePath string
			expected string
		}{
			{"/home/user/documents/file.txt", "/home/user/documents"},
			{"/home/user/downloads/file.pdf", "/home/user/downloads"},
			{"./file.txt", "(current directory)"},
			{"file.txt", "(current directory)"},
			{"/root/file.txt", "/root"},
		}
	}

	for _, test := range tests {
		result := extractDirectory(test.filePath)
		if result != test.expected {
			t.Errorf("extractDirectory(%s) = %s; want %s", test.filePath, result, test.expected)
		}
	}
}

func TestAnalyzeDirectories(t *testing.T) {
	// Create test duplicate groups with platform-appropriate paths
	var downloadsPath, documentsPath, desktopPath string
	if runtime.GOOS == "windows" {
		downloadsPath = "C:\\Users\\John\\Downloads"
		documentsPath = "C:\\Users\\John\\Documents"
		desktopPath = "C:\\Users\\John\\Desktop"
	} else {
		downloadsPath = "/home/downloads"
		documentsPath = "/home/documents"
		desktopPath = "/home/desktop"
	}

	groups := []scanner.DuplicateGroup{
		{
			Hash:  123,
			Files: []string{filepath.Join(downloadsPath, "file1.txt"), filepath.Join(documentsPath, "file1.txt")},
			Size:  100,
		},
		{
			Hash:  456,
			Files: []string{filepath.Join(downloadsPath, "file2.txt"), filepath.Join(downloadsPath, "file2_copy.txt"), filepath.Join(desktopPath, "file2.txt")},
			Size:  200,
		},
	}

	dirStats := analyzeDirectories(groups)

	// Test that we have the expected directories
	expectedDirs := []string{downloadsPath, documentsPath, desktopPath}
	if len(dirStats) != len(expectedDirs) {
		t.Errorf("Expected %d directories, got %d", len(expectedDirs), len(dirStats))
	}

	// Test downloads directory stats
	downloads := dirStats[downloadsPath]
	if downloads == nil {
		t.Error("Downloads directory not found in stats")
	} else {
		// Downloads has 1 duplicate from first group (2 files - 1 to keep) + 2 duplicates from second group (3 files - 1 to keep) = 3 duplicates
		if downloads.Count != 3 {
			t.Errorf("Downloads count = %d; want 3", downloads.Count)
		}
		// Size should be 1*100 + 2*200 = 500
		if downloads.TotalSize != 500 {
			t.Errorf("Downloads total size = %d; want 500", downloads.TotalSize)
		}
		// Should have 2 groups
		if len(downloads.Groups) != 2 {
			t.Errorf("Downloads groups = %d; want 2", len(downloads.Groups))
		}
	}

	// Test documents directory stats
	documents := dirStats[documentsPath]
	if documents == nil {
		t.Error("Documents directory not found in stats")
	} else {
		// Documents has 1 duplicate from first group
		if documents.Count != 1 {
			t.Errorf("Documents count = %d; want 1", documents.Count)
		}
		// Size should be 1*100 = 100
		if documents.TotalSize != 100 {
			t.Errorf("Documents total size = %d; want 100", documents.TotalSize)
		}
		// Should have 1 group
		if len(documents.Groups) != 1 {
			t.Errorf("Documents groups = %d; want 1", len(documents.Groups))
		}
	}

	// Test desktop directory stats
	desktop := dirStats[desktopPath]
	if desktop == nil {
		t.Error("Desktop directory not found in stats")
	} else {
		// Desktop has 2 duplicates from second group (3 files total - 1 to keep = 2 duplicates)
		if desktop.Count != 2 {
			t.Errorf("Desktop count = %d; want 2", desktop.Count)
		}
		// Size should be 2*200 = 400
		if desktop.TotalSize != 400 {
			t.Errorf("Desktop total size = %d; want 400", desktop.TotalSize)
		}
		// Should have 1 group
		if len(desktop.Groups) != 1 {
			t.Errorf("Desktop groups = %d; want 1", len(desktop.Groups))
		}
	}
}

func TestAnalyzeDirectoriesNoDuplicates(t *testing.T) {
	// Test with no duplicate groups
	groups := []scanner.DuplicateGroup{}
	dirStats := analyzeDirectories(groups)

	if len(dirStats) != 0 {
		t.Errorf("Expected 0 directories for no duplicates, got %d", len(dirStats))
	}
}

func TestAnalyzeDirectoriesSingleFiles(t *testing.T) {
	// Test with groups that have only single files (should be ignored)
	groups := []scanner.DuplicateGroup{
		{
			Hash:  123,
			Files: []string{"/home/documents/file1.txt"}, // Only 1 file, not a duplicate
			Size:  100,
		},
	}

	dirStats := analyzeDirectories(groups)

	if len(dirStats) != 0 {
		t.Errorf("Expected 0 directories for single files, got %d", len(dirStats))
	}
}