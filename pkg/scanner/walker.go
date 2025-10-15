package scanner

import (
	"io/fs"
	"path/filepath"
)

type FileInfo struct {
	Path    string
	Size    int64
	ModTime int64
}

func Walk(root string) ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		files = append(files, FileInfo{
			Path:    path,
			Size:    info.Size(),
			ModTime: info.ModTime().Unix(),
		})

		return nil
	})

	return files, err
}
