package hasher

import (
	"fmt"
	"io"
	"os"

	"github.com/cespare/xxhash/v2"
)

type QuickHash struct {
	Size     int64
	ModTime  int64
	HeadHash uint64
}

const headSize = 8 * 1024

func ComputeQuickHash(path string, size int64, modTime int64) (QuickHash, error) {
	f, err := os.Open(path)
	if err != nil {
		return QuickHash{}, err
	}
	defer f.Close()

	buf := make([]byte, headSize)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return QuickHash{}, err
	}

	h := xxhash.New()
	h.Write(buf[:n])

	return QuickHash{
		Size:     size,
		ModTime:  modTime,
		HeadHash: h.Sum64(),
	}, nil
}

func ComputeFullHash(path string) (uint64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	h := xxhash.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return 0, err
	}

	return h.Sum64(), nil
}

func (q QuickHash) String() string {
	return fmt.Sprintf("%d-%d-%x", q.Size, q.ModTime, q.HeadHash)
}
