package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseTime(t *testing.T) {
	tests := []struct {
		input       string
		expected    time.Time
		expectError bool
	}{
		{"1632992395.123456", time.Unix(1632992395, 123456), false},
		{"0.0", time.Unix(0, 0), false},
		{"invalid", time.Time{}, true},
	}

	for _, test := range tests {
		result, err := ParseTime(test.input)
		if test.expectError {
			assert.Error(t, err, "Expected error for input %s", test.input)
		} else {
			assert.NoError(t, err, "Unexpected error for input %s", test.input)
			assert.Equal(t, test.expected, result, "For input %s, expected %v, but got %v", test.input, test.expected, result)
		}
	}
}
