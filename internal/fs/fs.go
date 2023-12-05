// Package fs provides functionality to interact with the file system.
package fs

import "os"

// FileSystem is an interface that defines methods for interacting with the file system.
type FileSystem interface {
	// UserHomeDir returns the path to the current user's home directory.
	// It returns an error if the home directory cannot be determined.
	UserHomeDir() (string, error)

	// MkdirAll creates a directory named path, along with any necessary parents.
	// The permission bits perm are used for all directories that MkdirAll creates.
	// If path is already a directory, MkdirAll does nothing and returns nil.
	MkdirAll(path string, perm os.FileMode) error
}

// OSFileSystem is a struct that implements the FileSystem interface.
type OSFileSystem struct{}

// UserHomeDir is a method on the OSFileSystem struct that retrieves the path to the current user's home directory.
// It uses the os package's UserHomeDir function to do this.
// It returns an error if the home directory cannot be determined.
func (fs OSFileSystem) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

// MkdirAll is a method on the OSFileSystem struct that creates a directory named path, along with any necessary parents.
// It uses the os package's MkdirAll function to do this.
// The permission bits perm are used for all directories that MkdirAll creates.
// If path is already a directory, MkdirAll does nothing and returns nil.
func (fs OSFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}
