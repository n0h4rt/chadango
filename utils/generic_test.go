package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMin(t *testing.T) {
	assert.Equal(t, 1, Min(1, 3, 2), "Min should return the smallest number")
	assert.Equal(t, 0, Min(0, 5, 10), "Min should return the smallest number")
}

func TestMax(t *testing.T) {
	assert.Equal(t, 3, Max(1, 3, 2), "Max should return the largest number")
	assert.Equal(t, 10, Max(0, 5, 10), "Max should return the largest number")
}

func TestContains(t *testing.T) {
	arr := []int{1, 2, 3, 4, 5}
	assert.True(t, Contains(arr, 3), "Contains should return true if the item is found")
	assert.False(t, Contains(arr, 6), "Contains should return false if the item is not found")
}

func TestRemove(t *testing.T) {
	arr := []int{1, 2, 3, 4, 5}
	expectedArr := []int{1, 2, 4, 5}
	assert.Equal(t, expectedArr, Remove(arr, 3), "Remove should return the array without the removed item")
}
