package chadango

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/texttheater/golang-levenshtein/levenshtein"
)

// GetAnonName generates an anonymous name using the provided seed and sid.
func GetAnonName(seed, sid string) (name string) {
	if seed == "" {
		seed = "3452"
	}
	var (
		_seed int64
		_sid  int64
		sum   string
	)
	name = "anon"
	if len(sid) >= 8 {
		sid = sid[4:8]
	}
	for i := 0; i < len(sid); i++ {
		_seed, _ = strconv.ParseInt(seed[i:i+1], 10, 64)
		_sid, _ = strconv.ParseInt(sid[i:i+1], 10, 64)
		sum = fmt.Sprintf("%d", _seed+_sid)
		name += string(sum[len(sum)-1])
	}
	return
}

// CreateAnonSeed creates an anonymous seed using the provided name and sid.
func CreateAnonSeed(name, sid string) (seed string) {
	var (
		nInt int64
		sInt int64
	)
	if strings.HasPrefix(name, "anon") {
		name = name[4:Min(8, len(name))]
	}
	if len(sid) >= 8 {
		sid = sid[4:8]
	}
	if IsDigit(name) {
		for i := 0; i < len(name); i++ {
			nInt, _ = strconv.ParseInt(name[i:i+1], 10, 64)
			sInt, _ = strconv.ParseInt(sid[i:i+1], 10, 64)
			if nInt < sInt {
				nInt += 10
			}
			seed += fmt.Sprintf("%d", nInt-sInt)
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

// GetServer retrieves the server URL for the specified name.
// This function takes a string as input and returns a URL string for a server based on the input string.
// The returned URL is used for establishing a WebSocket connection.
// The function uses a weighted round-robin algorithm to select a server based on the input string.
// The input string is first converted to base36 integer by replacing "_" and "-" characters with "q".
// The first and second halves of the modified string are used to calculate a modulus ratio.
// The modulus ratio is then used to select a server based on its weight.
// The function returns a default URL if no server is selected.
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

// GenerateRandomString generates a random string of the specified length.
func GenerateRandomString(length int) string {
	charsetLength := int64(len(Charset))
	b := make([]byte, length)
	var n *big.Int
	for i := range b {
		n, _ = rand.Int(rand.Reader, big.NewInt(charsetLength))
		b[i] = Charset[n.Int64()]
	}
	return string(b)
}

// FindClosestMatch finds the closest matching string in the source array for the target string.
func FindClosestMatch(source []string, target string) (match string, ratio float64) {
	ratio = -1.0
	for _, src := range source {
		rat := levenshtein.RatioForStrings([]rune(src), []rune(target), levenshtein.DefaultOptions)
		if rat > ratio {
			ratio = rat
			match = src
		}
	}
	return
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

// CopyMap is a generic function that copies the source map to the destination map.
func CopyMap[K comparable, V any](src, dst map[K]V) {
	if dst == nil {
		dst = make(map[K]V)
	}
	for k, v := range src {
		dst[k] = v
	}
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
