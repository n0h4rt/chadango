package chadango

import (
	"bytes"
	"encoding/gob"
	"sync"
)

// SyncMap is a synchronized map that can be accessed concurrently.
//
// It provides thread-safe methods for setting, getting, deleting, and iterating over key-value pairs.
// The map also supports Gob encoding and decoding for persistence.
type SyncMap[K comparable, V any] struct {
	sync.RWMutex
	M map[K]V
}

// Set adds or updates a key-value pair in the SyncMap.
//
// Args:
//   - key: The key to add or update.
//   - val: The value to associate with the key.
func (sm *SyncMap[K, V]) Set(key K, val V) {
	sm.Lock()
	defer sm.Unlock()
	sm.M[key] = val
}

// Get retrieves the value associated with the specified key from the SyncMap.
//
// Args:
//   - key: The key to retrieve the value for.
//
// Returns:
//   - V: The value associated with the key.
//   - bool: True if the key exists in the map, false otherwise.
func (sm *SyncMap[K, V]) Get(key K) (val V, ok bool) {
	sm.RLock()
	defer sm.RUnlock()

	val, ok = sm.M[key]

	return
}

// Del removes the key-value pair with the specified key from the SyncMap.
//
// Args:
//   - key: The key to remove.
func (sm *SyncMap[K, V]) Del(key K) {
	sm.Lock()
	defer sm.Unlock()

	delete(sm.M, key)
}

// Len returns the number of key-value pairs in the SyncMap.
//
// Returns:
//   - int: The number of key-value pairs in the map.
func (sm *SyncMap[K, V]) Len() int {
	sm.RLock()
	defer sm.RUnlock()

	return len(sm.M)
}

// Range iterates over each key-value pair in the SyncMap and calls the specified function.
//
// Args:
//   - fun: The function to call for each key-value pair.
//
// The function takes two arguments:
//   - key: The key of the current key-value pair.
//   - value: The value of the current key-value pair.
//
// If the function returns false, the iteration stops.
func (sm *SyncMap[K, V]) Range(fun func(K, V) bool) {
	sm.RLock()
	defer sm.RUnlock()

	for k, v := range sm.M {
		if !fun(k, v) {
			return
		}
	}
}

// Clear removes all key-value pairs from the SyncMap.
func (sm *SyncMap[K, V]) Clear() {
	sm.Lock()
	defer sm.Unlock()

	sm.M = make(map[K]V)
}

// Keys returns a slice of keys in the SyncMap.
//
// Returns:
//   - []K: A slice containing all the keys in the map.
func (sm *SyncMap[K, V]) Keys() (keys []K) {
	sm.RLock()
	defer sm.RUnlock()

	for k := range sm.M {
		keys = append(keys, k)
	}

	return
}

// GobEncode encodes the SyncMap using Gob encoding.
//
// Returns:
//   - []byte: The encoded byte slice.
//   - error: An error if encoding fails.
func (sm *SyncMap[K, V]) GobEncode() ([]byte, error) {
	sm.RLock()
	defer sm.RUnlock()

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(sm.M)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GobDecode decodes the SyncMap using Gob decoding.
//
// Args:
//   - data: The encoded byte slice to decode.
//
// Returns:
//   - error: An error if decoding fails.
func (sm *SyncMap[K, V]) GobDecode(data []byte) error {
	sm.Lock()
	defer sm.Unlock()

	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&sm.M)

	if err != nil {
		return err
	}

	return nil
}

// NewSyncMap creates a new instance of SyncMap.
//
// Returns:
//   - SyncMap[K, V]: A new instance of SyncMap.
func NewSyncMap[K comparable, V any]() SyncMap[K, V] {
	return SyncMap[K, V]{M: map[K]V{}}
}

// OrderedSyncMap is a synchronized map that maintains the order of keys.
//
// It provides thread-safe methods for setting, getting, deleting, and iterating over key-value pairs in the order they were added.
// The map also supports Gob encoding and decoding for persistence.
type OrderedSyncMap[K comparable, V any] struct {
	sync.RWMutex
	K []K
	M map[K]V
}

// Set adds or updates a key-value pair in the OrderedSyncMap.
//
// Args:
//   - key: The key to add or update.
//   - val: The value to associate with the key.
func (sm *OrderedSyncMap[K, V]) Set(key K, val V) {
	sm.Lock()
	defer sm.Unlock()

	sm.del(key)
	sm.K = append(sm.K, key)
	sm.M[key] = val
}

// Get retrieves the value associated with the specified key from the OrderedSyncMap.
//
// Args:
//   - key: The key to retrieve the value for.
//
// Returns:
//   - V: The value associated with the key.
//   - bool: True if the key exists in the map, false otherwise.
func (sm *OrderedSyncMap[K, V]) Get(key K) (val V, ok bool) {
	sm.RLock()
	defer sm.RUnlock()

	val, ok = sm.M[key]

	return
}

// Del removes the key-value pair with the specified key from the OrderedSyncMap.
//
// Args:
//   - key: The key to remove.
func (sm *OrderedSyncMap[K, V]) Del(key K) {
	sm.Lock()
	defer sm.Unlock()

	sm.del(key)
}

// del removes the key from the OrderedSyncMap's key slice and deletes the corresponding value.
// This is not protected by Mutex, so keep it for internal use.
func (sm *OrderedSyncMap[K, V]) del(key K) {
	index := -1
	for i, k := range sm.K {
		if k == key {
			index = i
			break
		}
	}

	if index < 0 {
		return
	}

	sm.K = append(sm.K[:index], sm.K[index+1:]...)
	delete(sm.M, key)
}

// Len returns the number of key-value pairs in the OrderedSyncMap.
//
// Returns:
//   - int: The number of key-value pairs in the map.
func (sm *OrderedSyncMap[K, V]) Len() int {
	sm.RLock()
	defer sm.RUnlock()

	return len(sm.M)
}

// Range iterates over each key-value pair in the OrderedSyncMap and calls the specified function.
//
// Args:
//   - fun: The function to call for each key-value pair.
//
// The function takes two arguments:
//   - key: The key of the current key-value pair.
//   - value: The value of the current key-value pair.
//
// If the function returns false, the iteration stops.
func (sm *OrderedSyncMap[K, V]) Range(fun func(K, V) bool) {
	sm.RLock()
	defer sm.RUnlock()

	for _, k := range sm.K {
		if !fun(k, sm.M[k]) {
			return
		}
	}
}

// RangeReversed iterates over each key-value pair in the OrderedSyncMap in reverse order and calls the specified function.
//
// Args:
//   - fun: The function to call for each key-value pair.
//
// The function takes two arguments:
//   - key: The key of the current key-value pair.
//   - value: The value of the current key-value pair.
//
// If the function returns false, the iteration stops.
func (sm *OrderedSyncMap[K, V]) RangeReversed(fun func(K, V) bool) {
	sm.RLock()
	defer sm.RUnlock()

	for i := len(sm.K) - 1; i >= 0; i-- {
		if !fun(sm.K[i], sm.M[sm.K[i]]) {
			return
		}
	}
}

// Clear removes all key-value pairs from the OrderedSyncMap.
func (sm *OrderedSyncMap[K, V]) Clear() {
	sm.Lock()
	defer sm.Unlock()

	sm.K = make([]K, 0)
	sm.M = make(map[K]V)
}

// Keys returns a slice of keys in the OrderedSyncMap.
//
// Returns:
//   - []K: A slice containing all the keys in the map.
func (sm *OrderedSyncMap[K, V]) Keys() []K {
	sm.RLock()
	defer sm.RUnlock()

	return sm.K
}

// SetFront adds or updates a key-value pair at the front of the OrderedSyncMap.
//
// Args:
//   - key: The key to add or update.
//   - val: The value to associate with the key.
func (sm *OrderedSyncMap[K, V]) SetFront(key K, val V) {
	sm.Lock()
	defer sm.Unlock()

	sm.del(key)

	temp := make([]K, len(sm.K)+1)
	copy(temp[1:], sm.K)
	temp[0] = key

	sm.K = temp
	sm.M[key] = val
}

// TrimFront trims the front part of the OrderedSyncMap to the specified length.
//
// Args:
//   - length: The desired length of the map after trimming.
//
// It modifies the OrderedSyncMap in place,
// removing elements from the beginning of the map until the desired length is reached.
func (sm *OrderedSyncMap[K, V]) TrimFront(length int) {
	l := sm.Len()
	if l <= length {
		return
	}

	sm.Lock()
	defer sm.Unlock()

	trim := sm.K[:l-length]
	for _, k := range trim {
		sm.del(k)
	}
}

// GobEncode encodes the OrderedSyncMap using Gob.
//
// Returns:
//   - []byte: The encoded byte slice.
//   - error: An error if encoding fails.
func (sm *OrderedSyncMap[K, V]) GobEncode() ([]byte, error) {
	sm.RLock()
	defer sm.RUnlock()

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(sm.K)
	if err != nil {
		return nil, err
	}

	err = enc.Encode(sm.M)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GobDecode decodes the OrderedSyncMap using Gob.
//
// Args:
//   - data: The encoded byte slice to decode.
//
// Returns:
//   - error: An error if decoding fails.
func (sm *OrderedSyncMap[K, V]) GobDecode(data []byte) error {
	sm.Lock()
	defer sm.Unlock()

	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&sm.K)
	if err != nil {
		return err
	}

	err = dec.Decode(&sm.M)
	if err != nil {
		return err
	}

	return nil
}

// NewOrderedSyncMap creates a new instance of OrderedSyncMap.
//
// Returns:
//   - OrderedSyncMap[K, V]: A new instance of OrderedSyncMap.
func NewOrderedSyncMap[K comparable, V any]() OrderedSyncMap[K, V] {
	return OrderedSyncMap[K, V]{K: []K{}, M: map[K]V{}}
}
