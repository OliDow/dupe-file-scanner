package scanner

import (
	"io/fs"
	"path/filepath"
	"strings"
)

type FileInfo struct {
	Path    string
	Size    int64
	ModTime int64
}

var imageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".heic": true,
	".heif": true,
	".webp": true,
	".bmp":  true,
}

func isImage(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return imageExtensions[ext]
}

func Walk(root string, onlyImages bool) ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			return nil
		}

		if onlyImages && !isImage(path) {
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
