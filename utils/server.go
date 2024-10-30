package utils

import (
	"fmt"
	"strconv"
	"strings"
)

var (
	ctssw = map[string]float64{
		"sv10": 110, "sv12": 116, "sv8": 101, "sv6": 104, "sv4": 110, "sv2": 95, "sv0": 75,
	}
	ctssm = [][2]string{
		{"5", "sv0"}, {"6", "sv0"}, {"7", "sv0"}, {"8", "sv0"}, {"16", "sv0"},
		{"17", "sv0"}, {"18", "sv0"}, {"9", "sv2"}, {"11", "sv2"}, {"12", "sv2"},
		{"13", "sv2"}, {"14", "sv2"}, {"15", "sv2"}, {"19", "sv4"}, {"23", "sv4"},
		{"24", "sv4"}, {"25", "sv4"}, {"26", "sv4"}, {"28", "sv6"}, {"29", "sv6"},
		{"30", "sv6"}, {"31", "sv6"}, {"32", "sv6"}, {"33", "sv6"}, {"35", "sv8"},
		{"36", "sv8"}, {"37", "sv8"}, {"38", "sv8"}, {"39", "sv8"}, {"40", "sv8"},
		{"41", "sv8"}, {"42", "sv8"}, {"43", "sv8"}, {"44", "sv8"}, {"45", "sv8"},
		{"46", "sv8"}, {"47", "sv8"}, {"48", "sv8"}, {"49", "sv8"}, {"50", "sv8"},
		{"52", "sv10"}, {"53", "sv10"}, {"55", "sv10"}, {"57", "sv10"}, {"58", "sv10"},
		{"59", "sv10"}, {"60", "sv10"}, {"61", "sv10"}, {"62", "sv10"}, {"63", "sv10"},
		{"64", "sv10"}, {"65", "sv10"}, {"66", "sv10"}, {"68", "sv2"}, {"71", "sv12"},
		{"72", "sv12"}, {"73", "sv12"}, {"74", "sv12"}, {"75", "sv12"}, {"76", "sv12"},
		{"77", "sv12"}, {"78", "sv12"}, {"79", "sv12"}, {"80", "sv12"}, {"81", "sv12"},
		{"82", "sv12"}, {"83", "sv12"}, {"84", "sv12"},
	}
)

// GetServer returns the server URL for a given name.
//
// The function uses a weighted round-robin algorithm to select a server based on a calculated modulus ratio.
// It takes the name as input and calculates a ratio based on the first and second halves of the name.
// This ratio is then used to determine the server with the highest weight.
//
// Args:
//   - name: The name of the chat room.
//
// Returns:
//   - string: The server URL for the given name.
func GetServer(name string) string {
	var (
		firstHalf   int64
		secondHalf  int64 = 1000
		totalWeight float64
		serverEntry [2]string
		weightRatio float64
	)
	name = strings.ReplaceAll(name, "_", "q")
	name = strings.ReplaceAll(name, "-", "q")
	nameLen := len(name)
	firstHalf, _ = strconv.ParseInt(name[:Min(5, nameLen)], 36, 64)

	if nameLen > 5 {
		temp, _ := strconv.ParseInt(name[6:Min(9, nameLen)], 36, 64)
		secondHalf = Max(1000, temp)
	}

	modRatio := float64(firstHalf%secondHalf) / float64(secondHalf)

	// This can be pre-computed.
	for _, serverEntry = range ctssm {
		totalWeight += ctssw[serverEntry[1]]
	}

	for _, serverEntry = range ctssm {
		weightRatio += ctssw[serverEntry[1]] / totalWeight
		if modRatio <= weightRatio {
			return fmt.Sprintf("ws://s%s.chatango.com:8080/", serverEntry[0])
		}
	}

	return "ws://s5.chatango.com:8080/" // Default
}
