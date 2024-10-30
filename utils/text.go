package utils

import (
	"strconv"
	"strings"
)

// IsDigit checks whether the provided string represents a digit.
//
// The function attempts to parse the string as an integer.
// If the parsing is successful, it returns true, otherwise false.
//
// Args:
//   - strnum: The string to check.
//
// Returns:
//   - bool: True if the string represents a digit, otherwise false.
func IsDigit(strnum string) bool {
	_, err := strconv.ParseInt(strnum, 10, 64)
	return err == nil
}

// SplitTextIntoChunks splits the provided text into chunks of the specified size.
//
// The function splits the text into chunks of the specified size, ensuring that words are not split across chunks.
// It returns a slice of strings representing the chunks.
//
// Args:
//   - text: The text to split into chunks.
//   - chunkSize: The maximum size of each chunk.
//
// Returns:
//   - []string: A slice of strings representing the chunks.
func SplitTextIntoChunks(text string, chunkSize int) (chunks []string) {
	var currentChunk string
	var currentSize, wordSize int

	for _, word := range strings.Fields(text) {
		wordSize = len(word) + 1 // Include space after the word
		if currentSize+wordSize > chunkSize {
			// Start a new chunk
			chunks = append(chunks, currentChunk[:currentSize-1])
			currentChunk = ""
			currentSize = 0
		}

		currentChunk += word + " "
		currentSize += wordSize
	}

	if currentChunk != "" {
		chunks = append(chunks, currentChunk[:currentSize-1])
	}

	return
}
