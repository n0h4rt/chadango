package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeFlagChanges(t *testing.T) {
	tests := []struct {
		oldFlags    int64
		newFlags    int64
		expectedAdd int64
		expectedRem int64
	}{
		{3, 5, 4, 2},
		{10, 3, 1, 8},
		{0, 7, 7, 0},
	}

	for _, test := range tests {
		addedFlags, removedFlags := ComputeFlagChanges(test.oldFlags, test.newFlags)
		assert.Equal(t, test.expectedAdd, addedFlags, "ComputeFlagChanges addedFlags should match the expected result")
		assert.Equal(t, test.expectedRem, removedFlags, "ComputeFlagChanges removedFlags should match the expected result")
	}
}
