package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitTextIntoChunks(t *testing.T) {
	text := "Lorem ipsum dolor sit amet, consectetur adipiscing elit."

	// Chunk size: 12
	chunkSize := 12

	// Expected result: ["Lorem ipsum", "dolor sit", "amet,", "consectetur", "adipiscing", "elit."]
	expectedResult := []string{"Lorem ipsum", "dolor sit", "amet,", "consectetur", "adipiscing", "elit."}

	// Call the SplitTextIntoChunks function
	result := SplitTextIntoChunks(text, chunkSize)

	// Check if the result matches the expected result
	assert.Equal(t, expectedResult, result, "SplitTextIntoChunks result should match the expected result")
}
