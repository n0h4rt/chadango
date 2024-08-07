package chadango

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// GetAnonName generates an anonymous name using the provided seed and sid.
func GetAnonName(seed, sid int) (name string) {
	if seed == 0 {
		seed = 3452
	}
	var digit1, digit2, sum int

	seed %= 1e4
	sid %= 1e4
	result := 0

	for i := 1; seed > 0 || sid > 0; i *= 10 {
		digit1 = sid % 10
		digit2 = seed % 10
		sum = digit1 + digit2
		result = (sum%10)*i + result
		sid /= 10
		seed /= 10
	}

	return "anon" + strconv.Itoa(result)
}

// CreateAnonSeed creates an anonymous seed using the provided name and sid.
func CreateAnonSeed(name string, sid int) (seed int) {
	if strings.HasPrefix(name, "anon") {
		name = name[4:Min(8, len(name))]
	}
	sid %= 1e4
	if IsDigit(name) {
		var digit1, digit2 int
		mult := 1000

		for i := 0; i < len(name); i++ {
			digit1 = int(name[i] - '0')
			digit2 = sid / mult % 10

			if digit1 < digit2 {
				digit1 += 10
			}

			seed += (digit1 - digit2) * mult
			mult /= 10
		}
	}

	return
}

var (
	ctssw = map[string]int{
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
// It uses a weighted round-robin algorithm to select a server based on a calculated modulus ratio.
// If no server is selected, a default URL is returned.
func GetServer(name string) string {
	var (
		totalWeight float64
		serverEntry [2]string
		weightRatio float64
	)
	name = strings.ReplaceAll(name, "_", "q")
	name = strings.ReplaceAll(name, "-", "q")
	nameLen := len(name)
	temp, _ := strconv.ParseInt(name[:Min(5, nameLen)], 36, 64)
	firstHalf := int(temp)
	secondHalf := 1000

	if nameLen > 5 {
		temp, _ = strconv.ParseInt(name[6:Min(9, nameLen)], 36, 64)
		secondHalf = Max(1000, int(temp))
	}

	modRatio := float64(firstHalf%secondHalf) / float64(secondHalf)

	// This can be pre-computed.
	for _, serverEntry = range ctssm {
		totalWeight += float64(ctssw[serverEntry[1]])
	}

	for _, serverEntry = range ctssm {
		weightRatio += float64(ctssw[serverEntry[1]]) / totalWeight
		if modRatio <= weightRatio {
			return fmt.Sprintf("ws://s%s.chatango.com:8080/", serverEntry[0])
		}
	}

	return "ws://s5.chatango.com:8080/" // Default
}

// Contains is a generic function that checks whether the specified item is present in the given array.
func Contains[T comparable](arr []T, item T) bool {
	for _, i := range arr {
		if i == item {
			return true
		}
	}

	return false
}

// Remove is a generic function that removes the specified item from the given array.
func Remove[T comparable](arr []T, item T) []T {
	for i, v := range arr {
		if v == item {
			return append(arr[:i], arr[i+1:]...)
		}
	}

	return arr
}

// Min is a generic function that returns the minimum value among the provided numbers.
func Min[T Number](a T, b ...T) (c T) {
	c = a
	for _, n := range b {
		if n < c {
			c = n
		}
	}

	return
}

// Max is a generic function that returns the maximum value among the provided numbers.
func Max[T Number](a T, b ...T) (c T) {
	c = a
	for _, n := range b {
		if n > c {
			c = n
		}
	}

	return
}

// IsDigit checks whether the provided string represents a digit.
func IsDigit(strnum string) bool {
	_, err := strconv.ParseInt(strnum, 10, 64)
	return err == nil
}

// ParseTime parses the provided string representation of time into a time.Time value.
func ParseTime(strtime string) (t time.Time, err error) {
	var timeF float64
	if timeF, err = strconv.ParseFloat(strtime, 64); err == nil {
		t = time.Unix(0, int64(timeF*1e9))
	}

	return
}

// BoolZeroOrOne returns either "0" or "1"
func BoolZeroOrOne(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// FileExists checks whether the specified file exists.
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// SplitTextIntoChunks splits the provided text into chunks of the specified size.
func SplitTextIntoChunks(text string, chunkSize int) (chunks []string) {
	var currentChunk string
	var currentSize, wordSize int

	for _, word := range strings.Fields(text) {
		wordSize = len(word) + 1 // Include space after the word
		if currentSize+wordSize > chunkSize {
			// Start a new chunk
			chunks = append(chunks, currentChunk[:currentSize-1])
			currentChunk = ""
			currentSize = 0
		}

		currentChunk += word + " "
		currentSize += wordSize
	}

	if currentChunk != "" {
		chunks = append(chunks, currentChunk[:currentSize-1])
	}

	return
}

// ComputeFlagChanges computes the flag changes between the oldFlags and newFlags values.
func ComputeFlagChanges(oldFlags, newFlags int64) (addedFlags, removedFlags int64) {
	addedFlags = newFlags &^ oldFlags
	removedFlags = oldFlags &^ newFlags

	return
}

// LoadConfig loads the configuration from the specified file.
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config data: %v", err)
	}

	return &config, nil
}

// SaveConfig saves the configuration to the specified file.
func SaveConfig(filename string, config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config data: %v", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

func UsernameToURL(url, username string) string {
	path0 := username[0:1]
	path1 := username[0:1]
	if len(username) > 1 {
		path1 = username[1:2]
	}
	return fmt.Sprintf(url, path0, path1, username)
}
