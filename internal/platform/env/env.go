// Package env provides environment variable utilities.
package env

import "os"

// GetOrDefault returns the environment variable value or a default if unset/empty.
func GetOrDefault(key, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}

	return val
}
