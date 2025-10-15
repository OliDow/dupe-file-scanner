package scanner

import (
	"dupe-file-checker/pkg/hasher"
	"runtime"
	"sync"
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

func (s *Scanner) Scan(root string, onlyImages bool) ([]DuplicateGroup, error) {
	files, err := Walk(root, onlyImages)
	if err != nil {
		return nil, err
	}

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
				resultChan <- result{hash: fh, path: f.Path}
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
	for r := range resultChan {
		fullGroups[r.hash] = append(fullGroups[r.hash], r.path)
	}

	var duplicates []DuplicateGroup
	for hash, paths := range fullGroups {
		if len(paths) > 1 {
			duplicates = append(duplicates, DuplicateGroup{
				Hash:  hash,
				Files: paths,
			})
		}
	}

	return duplicates
}
