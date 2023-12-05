// Package env provides functionality to interact with the environment variables.
package env

import "os"

// Environment is an interface that defines a method for getting environment variables.
type Environment interface {
	// Getenv returns the value of the environment variable named by the key.
	// It returns an empty string if the key does not exist.
	Getenv(key string) string
}

// RealEnvironment is a struct that implements the Environment interface.
type RealEnvironment struct{}

// Getenv is a method on the RealEnvironment struct that retrieves the value of the environment variable named by the key.
// It uses the os package's Getenv function to do this.
// It returns an empty string if the key does not exist.
func (e RealEnvironment) Getenv(key string) string {
	return os.Getenv(key)
}
