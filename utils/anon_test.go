package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAnonName(t *testing.T) {
	tests := []struct {
		seed     int
		sid      int
		expected string
	}{
		{45621234, 66668888, "anon9012"},
		{0, 1234, "anon4686"},
		{3452, 5678, "anon8020"},
	}

	for _, test := range tests {
		result := GetAnonName(test.seed, test.sid)
		assert.Equal(t, test.expected, result, "GetAnonName result should match the expected result")
	}
}

func TestCreateAnonSeed(t *testing.T) {
	tests := []struct {
		name     string
		sid      int
		expected int
	}{
		{"anon9012", 66668888, 1234},
		{"anon1234", 1234, 0},
		{"nonanon", 3452, 0},
	}

	for _, test := range tests {
		result := CreateAnonSeed(test.name, test.sid)
		assert.Equal(t, test.expected, result, "CreateAnonSeed result should match the expected result")
	}
}
