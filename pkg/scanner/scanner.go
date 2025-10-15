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

func (s *Scanner) Scan(root string) ([]DuplicateGroup, error) {
	files, err := Walk(root)
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
	quickGroups := make(map[hasher.QuickHash][]FileInfo)
	var mu sync.Mutex

	var wg sync.WaitGroup
	workChan := make(chan FileInfo, 100)

	for i := 0; i < s.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range workChan {
				qh, err := hasher.ComputeQuickHash(f.Path, f.Size, f.ModTime)
				if err != nil {
					continue
				}

				mu.Lock()
				quickGroups[qh] = append(quickGroups[qh], f)
				mu.Unlock()
			}
		}()
	}

	for _, group := range sizeGroups {
		for _, f := range group {
			workChan <- f
		}
	}
	close(workChan)
	wg.Wait()

	filtered := make(map[hasher.QuickHash][]FileInfo)
	for qh, group := range quickGroups {
		if len(group) > 1 {
			filtered[qh] = group
		}
	}
	return filtered
}

func (s *Scanner) processFullHashes(quickGroups map[hasher.QuickHash][]FileInfo) []DuplicateGroup {
	fullGroups := make(map[uint64][]string)
	var mu sync.Mutex

	var wg sync.WaitGroup
	workChan := make(chan FileInfo, 100)

	for i := 0; i < s.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range workChan {
				fh, err := hasher.ComputeFullHash(f.Path)
				if err != nil {
					continue
				}

				mu.Lock()
				fullGroups[fh] = append(fullGroups[fh], f.Path)
				mu.Unlock()
			}
		}()
	}

	for _, group := range quickGroups {
		for _, f := range group {
			workChan <- f
		}
	}
	close(workChan)
	wg.Wait()

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
