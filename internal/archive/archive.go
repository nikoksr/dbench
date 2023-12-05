package archive

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver/v4"
)

const fileExtension = ".tar.zst"

func defaultFormat() archiver.CompressedArchive {
	return archiver.CompressedArchive{
		Archival:    archiver.Tar{},
		Compression: archiver.Zstd{},
	}
}

// IsArchive checks if the provided path is an archive file.
// It returns true if the file has the archive file extension, false otherwise.
func IsArchive(path string) bool {
	return strings.HasSuffix(path, fileExtension)
}

// Directory creates an archive from the files in the provided source path.
// The archive is created at the target path.
// The function returns the path to the created archive and any error encountered.
func Directory(ctx context.Context, sourcePath, targetPath string) (string, error) {
	// Read files from disk
	files, err := archiver.FilesFromDisk(nil, map[string]string{
		sourcePath: "", // Put files in the root of the archive
	})
	if err != nil {
		return "", fmt.Errorf("read files from disk: %w", err)
	}

	// Sanitize target path
	ext := filepath.Ext(targetPath)
	targetPath = strings.TrimSuffix(targetPath, ext)
	targetPath += fileExtension

	// Create output file
	out, err := os.Create(targetPath)
	if err != nil {
		return "", fmt.Errorf("create output file: %w", err)
	}
	defer out.Close()

	// Note: We can probably gain some performance by using format.ArchiveAsync(), but for the scope of this project
	// and simplicity, we'll stick to the synchronous version.
	return targetPath, defaultFormat().Archive(ctx, out, files)
}

// Extract extracts the files from the provided reader into the specified output directory.
// The function returns any error encountered.
func Extract(ctx context.Context, r io.Reader, outDir string) error {
	// Note: Adding a callback function to the Extract method would probably make this more efficient, but, again, for
	// the scope of this project and simplicity, we'll stick to the simple version and save all files to disk. This
	// also integrates better with the rest of the codebase.
	handler := func(ctx context.Context, f archiver.File) error {
		// Skip directories
		if f.IsDir() {
			return nil
		}
		// Skip non JSON files
		if !strings.HasSuffix(f.Name(), ".json") {
			return nil
		}

		// Create file
		target, err := os.Create(filepath.Join(outDir, f.Name()))
		if err != nil {
			return fmt.Errorf("create file: %w", err)
		}
		defer target.Close()

		source, err := f.Open()
		if err != nil {
			return fmt.Errorf("open file: %w", err)
		}
		defer source.Close()

		// Copy file contents
		if _, err := io.Copy(target, source); err != nil {
			return fmt.Errorf("copy file contents: %w", err)
		}

		return nil
	}

	return defaultFormat().Extract(ctx, r, nil, handler)
}
