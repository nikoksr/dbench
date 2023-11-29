package mocks

import (
	stdfs "io/fs"

	"github.com/stretchr/testify/mock"

	"github.com/nikoksr/dbench/internal/fs"
)

// MockFileSystem is a struct that implements the FileSystem interface from the fs package.
// It uses the testify/mock package to create mock implementations of the FileSystem methods.
type MockFileSystem struct {
	mock.Mock
	fs.FileSystem
}

// UserHomeDir is a method on the MockFileSystem struct that mocks the UserHomeDir method of the FileSystem interface.
// It uses the Called method from the testify/mock package to simulate the method call and return the mocked results.
func (m *MockFileSystem) UserHomeDir() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// MkdirAll is a method on the MockFileSystem struct that mocks the MkdirAll method of the FileSystem interface.
// It uses the Called method from the testify/mock package to simulate the method call and return the mocked results.
func (m *MockFileSystem) MkdirAll(path string, perm stdfs.FileMode) error {
	args := m.Called(path, perm)
	return args.Error(0)
}
