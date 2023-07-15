package chadango

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyncMap_SetAndGet(t *testing.T) {
	// Create a new instance of SyncMap
	sm := NewSyncMap[string, string]()

	// Set key-value pairs
	sm.Set("key1", "value1")
	sm.Set("key2", "value2")
	sm.Set("key3", "value3")

	// Get values for existing keys
	val1, ok1 := sm.Get("key1")
	val2, ok2 := sm.Get("key2")
	val3, ok3 := sm.Get("key3")

	// Assert that values are retrieved correctly
	assert.True(t, ok1)
	assert.Equal(t, "value1", val1)
	assert.True(t, ok2)
	assert.Equal(t, "value2", val2)
	assert.True(t, ok3)
	assert.Equal(t, "value3", val3)

	// Get value for non-existing key
	val4, ok4 := sm.Get("key4")

	// Assert that value is not found
	assert.False(t, ok4)
	assert.Equal(t, "", val4)
}

func TestSyncMap_Del(t *testing.T) {
	// Create a new instance of SyncMap
	sm := NewSyncMap[string, string]()

	// Set key-value pairs
	sm.Set("key1", "value1")
	sm.Set("key2", "value2")
	sm.Set("key3", "value3")

	// Delete a key-value pair
	sm.Del("key2")

	// Get values for existing keys
	val1, ok1 := sm.Get("key1")
	val2, ok2 := sm.Get("key2")
	val3, ok3 := sm.Get("key3")

	// Assert that deleted key is not found
	assert.True(t, ok1)
	assert.Equal(t, "value1", val1)
	assert.False(t, ok2)
	assert.Equal(t, "", val2)
	assert.True(t, ok3)
	assert.Equal(t, "value3", val3)
}

func TestSyncMap_Len(t *testing.T) {
	// Create a new instance of SyncMap
	sm := NewSyncMap[string, string]()

	// Set key-value pairs
	sm.Set("key1", "value1")
	sm.Set("key2", "value2")
	sm.Set("key3", "value3")

	// Get the length of the map
	length := sm.Len()

	// Assert that the length is correct
	assert.Equal(t, 3, length)
}

func TestSyncMap_Range(t *testing.T) {
	// Create a new instance of SyncMap
	sm := NewSyncMap[string, string]()

	// Set key-value pairs
	sm.Set("key1", "value1")
	sm.Set("key2", "value2")
	sm.Set("key3", "value3")

	// Iterate over the map and collect the keys and values
	keys := []string{}
	values := []string{}
	sm.Range(func(key string, value string) bool {
		keys = append(keys, key)
		values = append(values, value)
		return true
	})

	// Assert that the keys and values are collected correctly
	assert.ElementsMatch(t, []string{"key1", "key2", "key3"}, keys)
	assert.ElementsMatch(t, []string{"value1", "value2", "value3"}, values)
}

func TestSyncMap_Clear(t *testing.T) {
	// Create a new instance of SyncMap
	sm := NewSyncMap[string, string]()

	// Set key-value pairs
	sm.Set("key1", "value1")
	sm.Set("key2", "value2")
	sm.Set("key3", "value3")

	// Clear the map
	sm.Clear()

	// Get the length of the map
	length := sm.Len()

	// Assert that the map is empty
	assert.Equal(t, 0, length)
}

func TestSyncMap_Keys(t *testing.T) {
	// Create a new instance of SyncMap
	sm := NewSyncMap[string, string]()

	// Set key-value pairs
	sm.Set("key1", "value1")
	sm.Set("key2", "value2")
	sm.Set("key3", "value3")

	// Get the keys
	keys := sm.Keys()

	// Assert that the keys are retrieved correctly
	assert.ElementsMatch(t, []string{"key1", "key2", "key3"}, keys)
}

func TestOrderedSyncMap_SetAndGet(t *testing.T) {
	// Create a new instance of OrderedSyncMap
	sm := NewOrderedSyncMap[string, string]()

	// Set key-value pairs
	sm.Set("key1", "value1")
	sm.Set("key2", "value2")
	sm.Set("key3", "value3")

	// Get values for existing keys
	val1, ok1 := sm.Get("key1")
	val2, ok2 := sm.Get("key2")
	val3, ok3 := sm.Get("key3")

	// Assert that values are retrieved correctly
	assert.True(t, ok1)
	assert.Equal(t, "value1", val1)
	assert.True(t, ok2)
	assert.Equal(t, "value2", val2)
	assert.True(t, ok3)
	assert.Equal(t, "value3", val3)

	// Get value for non-existing key
	val4, ok4 := sm.Get("key4")

	// Assert that value is not found
	assert.False(t, ok4)
	assert.Equal(t, "", val4)
}

func TestOrderedSyncMap_Del(t *testing.T) {
	// Create a new instance of OrderedSyncMap
	sm := NewOrderedSyncMap[string, string]()

	// Set key-value pairs
	sm.Set("key1", "value1")
	sm.Set("key2", "value2")
	sm.Set("key3", "value3")

	// Delete a key-value pair
	sm.Del("key2")

	// Get values for existing keys
	val1, ok1 := sm.Get("key1")
	val2, ok2 := sm.Get("key2")
	val3, ok3 := sm.Get("key3")

	// Assert that deleted key is not found
	assert.True(t, ok1)
	assert.Equal(t, "value1", val1)
	assert.False(t, ok2)
	assert.Equal(t, "", val2)
	assert.True(t, ok3)
	assert.Equal(t, "value3", val3)
}

func TestOrderedSyncMap_Len(t *testing.T) {
	// Create a new instance of OrderedSyncMap
	sm := NewOrderedSyncMap[string, string]()

	// Set key-value pairs
	sm.Set("key1", "value1")
	sm.Set("key2", "value2")
	sm.Set("key3", "value3")

	// Get the length of the map
	length := sm.Len()

	// Assert that the length is correct
	assert.Equal(t, 3, length)
}

func TestOrderedSyncMap_Range(t *testing.T) {
	// Create a new instance of OrderedSyncMap
	sm := NewOrderedSyncMap[string, string]()

	// Set key-value pairs
	sm.Set("key1", "value1")
	sm.Set("key2", "value2")
	sm.Set("key3", "value3")

	// Iterate over the map and collect the keys and values
	keys := []string{}
	values := []string{}
	sm.Range(func(key string, value string) bool {
		keys = append(keys, key)
		values = append(values, value)
		return true
	})

	// Assert that the keys and values are ordered correctly
	assert.Equal(t, []string{"key1", "key2", "key3"}, keys)
	assert.Equal(t, []string{"value1", "value2", "value3"}, values)
}

func TestOrderedSyncMap_RangeReversed(t *testing.T) {
	// Create a new instance of OrderedSyncMap
	sm := NewOrderedSyncMap[string, string]()

	// Set key-value pairs
	sm.Set("key1", "value1")
	sm.Set("key2", "value2")
	sm.Set("key3", "value3")

	// Iterate over the map and collect the keys and values
	keys := []string{}
	values := []string{}
	sm.RangeReversed(func(key string, value string) bool {
		keys = append(keys, key)
		values = append(values, value)
		return true
	})

	// Assert that the keys and values are reversed correctly
	assert.Equal(t, []string{"key3", "key2", "key1"}, keys)
	assert.Equal(t, []string{"value3", "value2", "value1"}, values)
}

func TestOrderedSyncMap_Clear(t *testing.T) {
	// Create a new instance of OrderedSyncMap
	sm := NewOrderedSyncMap[string, string]()

	// Set key-value pairs
	sm.Set("key1", "value1")
	sm.Set("key2", "value2")
	sm.Set("key3", "value3")

	// Clear the map
	sm.Clear()

	// Get the length of the map
	length := sm.Len()

	// Assert that the map is empty
	assert.Equal(t, 0, length)
}

func TestOrderedSyncMap_Keys(t *testing.T) {
	// Create a new instance of OrderedSyncMap
	sm := NewOrderedSyncMap[string, string]()

	// Set key-value pairs
	sm.Set("key1", "value1")
	sm.Set("key2", "value2")
	sm.Set("key3", "value3")

	// Get the keys
	keys := sm.Keys()

	// Assert that the keys are retrieved correctly
	assert.Equal(t, []string{"key1", "key2", "key3"}, keys)
}

func TestOrderedSyncMap_SetFront(t *testing.T) {
	// Create a new instance of OrderedSyncMap
	sm := NewOrderedSyncMap[string, string]()

	// Set key-value pairs
	sm.Set("key1", "value1")
	sm.Set("key2", "value2")
	sm.Set("key3", "value3")

	// Set a key-value pair at the front
	sm.SetFront("key4", "value4")

	// Get the keys
	keys := sm.Keys()

	// Assert that the key-value pair is on order
	assert.Equal(t, []string{"key4", "key1", "key2", "key3"}, keys)
}

func TestOrderedSyncMap_TrimFront(t *testing.T) {
	// Create a new instance of OrderedSyncMap
	sm := NewOrderedSyncMap[string, string]()

	// Set key-value pairs
	sm.Set("key1", "value1")
	sm.Set("key2", "value2")
	sm.Set("key3", "value3")

	// Trim the map with length of 2
	sm.TrimFront(2)

	// Get the length of the map
	length := sm.Len()

	// Assert that the length is correct
	assert.Equal(t, 2, length)
}
