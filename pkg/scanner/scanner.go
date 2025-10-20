package scanner

import (
	"dupe-file-checker/pkg/hasher"
	"fmt"
	"runtime"
	"sync"
	"time"
)

type DuplicateGroup struct {
	Hash  uint64
	Files []string
	Size  int64
}

type Scanner struct {
	workers int
}

func New() *Scanner {
	return &Scanner{
		workers: runtime.NumCPU(),
	}
}

// estimateScanTime estimates scan duration based on benchmark data
// Benchmark data: 10 files = ~0.2ms, 100 files = ~1.6ms
func estimateScanTime(fileCount int) time.Duration {
	if fileCount <= 0 {
		return 0
	}

	// Use logarithmic scaling based on benchmark data points
	// For small file counts (< 10), use linear interpolation from 0.2ms
	if fileCount <= 10 {
		return time.Duration(float64(fileCount) * 0.02 * float64(time.Millisecond))
	}

	// For larger counts, use the relationship: time ≈ 0.02 * fileCount^1.1 ms
	// This accounts for the slight increase in complexity with more files
	timeMs := 0.02 * float64(fileCount) * (1.0 + float64(fileCount)/10000.0)

	return time.Duration(timeMs * float64(time.Millisecond))
}

func (s *Scanner) Scan(root string, onlyImages bool) ([]DuplicateGroup, error) {
	files, err := Walk(root, onlyImages)
	if err != nil {
		return nil, err
	}

	// Display the total count of included files and estimated time before starting scan
	fileType := "files"
	if onlyImages {
		fileType = "image files"
	}

	estimatedTime := estimateScanTime(len(files))
	var timeStr string
	if estimatedTime < time.Second {
		timeStr = fmt.Sprintf("%.1fms", float64(estimatedTime)/float64(time.Millisecond))
	} else if estimatedTime < time.Minute {
		timeStr = fmt.Sprintf("%.1fs", estimatedTime.Seconds())
	} else {
		timeStr = fmt.Sprintf("%.1fm", estimatedTime.Minutes())
	}

	fmt.Printf("Found %d %s to scan. Estimated time: %s\nStarting duplicate detection...\n\n", len(files), fileType, timeStr)

	sizeGroups := s.groupBySize(files)
	quickGroups := s.processQuickHashes(sizeGroups)
	duplicates := s.processFullHashes(quickGroups)

	return duplicates, nil
}

func (s *Scanner) groupBySize(files []FileInfo) map[int64][]FileInfo {
	groups := make(map[int64][]FileInfo)
	for _, f := range files {
		groups[f.Size] = append(groups[f.Size], f)
	}

	filtered := make(map[int64][]FileInfo)
	for size, group := range groups {
		if len(group) > 1 {
			filtered[size] = group
		}
	}
	return filtered
}

func (s *Scanner) processQuickHashes(sizeGroups map[int64][]FileInfo) map[hasher.QuickHash][]FileInfo {
	type result struct {
		hash hasher.QuickHash
		file FileInfo
	}

	workChan := make(chan FileInfo, 100)
	resultChan := make(chan result, 100)

	var wg sync.WaitGroup
	for i := 0; i < s.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range workChan {
				qh, err := hasher.ComputeQuickHash(f.Path, f.Size, f.ModTime)
				if err != nil {
					continue
				}
				resultChan <- result{hash: qh, file: f}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	go func() {
		for _, group := range sizeGroups {
			for _, f := range group {
				workChan <- f
			}
		}
		close(workChan)
	}()

	quickGroups := make(map[hasher.QuickHash][]FileInfo)
	for r := range resultChan {
		quickGroups[r.hash] = append(quickGroups[r.hash], r.file)
	}

	filtered := make(map[hasher.QuickHash][]FileInfo)
	for qh, group := range quickGroups {
		if len(group) > 1 {
			filtered[qh] = group
		}
	}
	return filtered
}

func (s *Scanner) processFullHashes(quickGroups map[hasher.QuickHash][]FileInfo) []DuplicateGroup {
	type result struct {
		hash uint64
		path string
		size int64
	}

	workChan := make(chan FileInfo, 100)
	resultChan := make(chan result, 100)

	var wg sync.WaitGroup
	for i := 0; i < s.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range workChan {
				fh, err := hasher.ComputeFullHash(f.Path)
				if err != nil {
					continue
				}
				resultChan <- result{hash: fh, path: f.Path, size: f.Size}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	go func() {
		for _, group := range quickGroups {
			for _, f := range group {
				workChan <- f
			}
		}
		close(workChan)
	}()

	fullGroups := make(map[uint64][]string)
	fileSizes := make(map[uint64]int64)
	for r := range resultChan {
		fullGroups[r.hash] = append(fullGroups[r.hash], r.path)
		fileSizes[r.hash] = r.size // All files with same hash have same size
	}

	var duplicates []DuplicateGroup
	for hash, paths := range fullGroups {
		if len(paths) > 1 {
			duplicates = append(duplicates, DuplicateGroup{
				Hash:  hash,
				Files: paths,
				Size:  fileSizes[hash],
			})
		}
	}

	return duplicates
}
