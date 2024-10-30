package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetServer(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"khususme", "ws://s39.chatango.com:8080/"},
		{"animeindofun", "ws://s50.chatango.com:8080/"},
		{"komikcastsite", "ws://s16.chatango.com:8080/"},
	}

	for _, test := range tests {
		result := GetServer(test.name)
		assert.Equal(t, test.expected, result, "GetServer result should match the expected result")
	}
}
