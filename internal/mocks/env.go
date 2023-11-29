package mocks

import (
	"github.com/stretchr/testify/mock"

	"github.com/nikoksr/dbench/internal/env"
)

// MockEnvironment is a struct that implements the Environment interface from the env package.
// It uses the testify/mock package to create mock implementations of the Environment methods.
type MockEnvironment struct {
	mock.Mock
	env.Environment
}

// Getenv is a method on the MockEnvironment struct that mocks the Getenv method of the Environment interface.
// It uses the Called method from the testify/mock package to simulate the method call and return the mocked results.
func (m *MockEnvironment) Getenv(key string) string {
	args := m.Called(key)
	return args.String(0)
}
