package chadango

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAnonName(t *testing.T) {
	seed := "1234"
	sid := "66668888"

	// Expected result: "anon9012"
	expectedResult := "anon9012"

	// Call the GetAnonName function
	result := GetAnonName(seed, sid)

	// Check if the result matches the expected result
	assert.Equal(t, expectedResult, result, "GetAnonName result should match the expected result")
}

func TestCreateAnonSeed(t *testing.T) {
	name := "anon9012"
	sid := "66668888"

	// Expected result: "1234"
	expectedResult := "1234"

	// Call the CreateAnonSeed function
	result := CreateAnonSeed(name, sid)

	// Check if the result matches the expected result
	assert.Equal(t, expectedResult, result, "CreateAnonSeed result should match the expected result")
}

func TestGetServer(t *testing.T) {
	//   - "khususme" -> ws://s39.chatango.com:8080/
	//   - "animeindofun" -> ws://s50.chatango.com:8080/
	//   - "komikcastsite" -> ws://s16.chatango.com:8080/
	name := "khususme"

	// Expected result: "ws://s39.chatango.com:8080/"
	expectedResult := "ws://s39.chatango.com:8080/"

	// Call the GetServer function
	result := GetServer(name)

	// Check if the result matches the expected result
	assert.Equal(t, expectedResult, result, "GetServer result should match the expected result")
}

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

func TestComputeFlagChanges(t *testing.T) {
	oldFlags := int64(0b01010010)
	newFlags := int64(0b00110010)

	// Expected results: addedFlags = bit 6, removedFlags = bit 7
	expectedAddedFlags := int64(0b00100000)
	expectedRemovedFlags := int64(0b01000000)

	// Call the ComputeFlagChanges function
	addedFlags, removedFlags := ComputeFlagChanges(oldFlags, newFlags)

	// Check if the results match the expected results
	assert.Equal(t, expectedAddedFlags, addedFlags, "ComputeFlagChanges addedFlags should match the expected result")
	assert.Equal(t, expectedRemovedFlags, removedFlags, "ComputeFlagChanges removedFlags should match the expected result")
}
