package reporter

import (
	"dupe-file-checker/pkg/scanner"
	"fmt"
	"path/filepath"
	"sort"
)

type DirectoryStats struct {
	Path       string
	Count      int
	TotalSize  int64
	Groups     []scanner.DuplicateGroup
}

// extractDirectory gets the directory path from a file path
func extractDirectory(filePath string) string {
	dir := filepath.Dir(filePath)
	if dir == "." {
		return "(current directory)"
	}
	return dir
}

// formatSize converts bytes to human-readable format
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// analyzeDirectories groups duplicates by directory and calculates statistics
func analyzeDirectories(groups []scanner.DuplicateGroup) map[string]*DirectoryStats {
	dirStats := make(map[string]*DirectoryStats)

	for _, group := range groups {
		if len(group.Files) < 2 {
			continue
		}

		// Track which directories this duplicate group affects
		affectedDirs := make(map[string]bool)
		for _, file := range group.Files {
			dir := extractDirectory(file)
			affectedDirs[dir] = true
		}

		// Add this group to each affected directory's stats
		for dir := range affectedDirs {
			if dirStats[dir] == nil {
				dirStats[dir] = &DirectoryStats{
					Path:   dir,
					Groups: make([]scanner.DuplicateGroup, 0),
				}
			}

			dirStats[dir].Count += len(group.Files) - 1 // subtract 1 because we keep one copy
			dirStats[dir].TotalSize += group.Size * int64(len(group.Files)-1)
			dirStats[dir].Groups = append(dirStats[dir].Groups, group)
		}
	}

	return dirStats
}

func PrintDuplicates(groups []scanner.DuplicateGroup) {
	if len(groups) == 0 {
		fmt.Println("No duplicates found")
		return
	}

	// Analyze directories
	dirStats := analyzeDirectories(groups)

	// Print directory summary header
	printDirectorySummary(dirStats)

	// Print detailed duplicates grouped by directory
	printDirectoryGroupedDuplicates(dirStats)
}

// printDirectorySummary prints the header with directory statistics
func printDirectorySummary(dirStats map[string]*DirectoryStats) {
	fmt.Println("📁 DUPLICATE SUMMARY BY DIRECTORY")

	// Convert map to slice for sorting
	var dirs []*DirectoryStats
	for _, stats := range dirStats {
		dirs = append(dirs, stats)
	}

	// Sort by total size (largest wasted space first)
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].TotalSize > dirs[j].TotalSize
	})

	var totalDuplicates int
	var totalSize int64

	for _, stats := range dirs {
		fmt.Printf("├─ %s   %d duplicates (%s wasted)\n",
			stats.Path, stats.Count, formatSize(stats.TotalSize))
		totalDuplicates += stats.Count
		totalSize += stats.TotalSize
	}

	fmt.Printf("\nTotal: %d duplicate files could save %s\n\n",
		totalDuplicates, formatSize(totalSize))
}

// printDirectoryGroupedDuplicates prints detailed file listings grouped by directory
func printDirectoryGroupedDuplicates(dirStats map[string]*DirectoryStats) {
	fmt.Println("📂 DUPLICATES BY DIRECTORY:\n")

	// Convert map to slice for consistent ordering
	var dirs []*DirectoryStats
	for _, stats := range dirStats {
		dirs = append(dirs, stats)
	}

	// Sort by total size (largest wasted space first)
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].TotalSize > dirs[j].TotalSize
	})

	for _, stats := range dirs {
		fmt.Printf("%s (%d duplicates):\n", stats.Path, stats.Count)

		// Group duplicates and show them
		groupNum := 1
		for _, group := range stats.Groups {
			if len(group.Files) < 2 {
				continue
			}

			// Extract filename for display
			filename := filepath.Base(group.Files[0])
			fmt.Printf("  Group %d: %s (%d copies, %s each)\n",
				groupNum, filename, len(group.Files), formatSize(group.Size))

			for _, file := range group.Files {
				fmt.Printf("    - %s\n", file)
			}
			fmt.Println()
			groupNum++
		}
		fmt.Println()
	}
}
