package utils

import "os"

// FileExists checks whether the specified file exists.
//
// Args:
//   - filename: The name of the file to check.
//
// Returns:
//   - bool: True if the file exists, otherwise false.
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
