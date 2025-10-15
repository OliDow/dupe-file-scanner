# Duplicate File Checker

Lightweight, efficient duplicate file detection tool written in Go.

## Features

- **3-Stage Hash Strategy** - Efficient filtering to minimize disk I/O
  - Stage 1: Size grouping (no I/O)
  - Stage 2: Quick hash (size + mtime + first 8KB)
  - Stage 3: Full file hash (only when necessary)
- **Concurrent Processing** - Worker pool using all CPU cores
- **Image Filtering** - Optional flag to scan only image files
- **xxHash Algorithm** - 10x faster than MD5 for non-cryptographic use
- **Minimal Memory** - ~200 bytes per file, ~20MB for 100K files

## Performance

- 10 files: ~0.2ms
- 100 files: ~1.6ms
- Memory efficient: ~1.7MB for 100 files

## Usage

```bash
# Build
go build -o dupe-checker

# Scan all files
./dupe-checker /path/to/scan

# Scan only image files
./dupe-checker --only-images /path/to/photos
```

## Supported Image Formats

When using `--only-images` flag:
- jpg, jpeg, png, gif
- heic, heif (Apple photos)
- webp (modern web)
- bmp (bitmap)

## Architecture

```
pkg/
├── scanner/     File traversal and orchestration
├── hasher/      Hash computation (quick + full)
├── grouper/     Duplicate detection logic
└── reporter/    Output formatting
```

## Testing

```bash
# Run all tests
go test ./...

# Run with verbose output
go test ./... -v

# Run benchmarks
go test -bench=. -benchmem ./pkg/scanner/
```

## How It Works

1. **Walk** - Recursively traverse directories using filepath.WalkDir
2. **Group by Size** - Fast filter, eliminates unique-sized files
3. **Quick Hash** - Compute hash of first 8KB + metadata (~95% accuracy)
4. **Full Hash** - Verify potential duplicates with complete file hash
5. **Report** - Display duplicate groups

## Technical Details

- **Hash Function**: xxHash (64-bit, non-cryptographic)
- **Concurrency**: Worker pool pattern with runtime.NumCPU() workers
- **File Properties**: Size, ModTime, Content Hash
- **False Positives**: Prevented by 3-stage verification
