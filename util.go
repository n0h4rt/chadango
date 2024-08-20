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
//
// The function uses a specific algorithm to generate an anonymous name based on the provided seed and sid values.
// It ensures that the generated name starts with "anon" followed by a unique numerical identifier.
//
// Args:
//   - seed: The seed value used for generating the anonymous name.
//   - sid: The session ID of the user.
//
// Returns:
//   - string: The generated anonymous name.
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
//
// The function generates an anonymous seed value based on the provided name and sid.
// It extracts the numerical part from the name (if it starts with "anon") and combines it with the sid to create a unique seed.
//
// Args:
//   - name: The name of the user.
//   - sid: The session ID of the user.
//
// Returns:
//   - int: The generated anonymous seed value.
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

// Contains is a generic function that checks whether the specified item is present in the given array.
//
// The function iterates through the array and compares each element with the specified item.
// It returns true if the item is found in the array, otherwise false.
//
// Args:
//   - arr: The array to search in.
//   - item: The item to search for.
//
// Returns:
//   - bool: True if the item is found in the array, otherwise false.
func Contains[T comparable](arr []T, item T) bool {
	for _, i := range arr {
		if i == item {
			return true
		}
	}

	return false
}

// Remove is a generic function that removes the specified item from the given array.
//
// The function iterates through the array and removes the first occurrence of the specified item.
// It returns the modified array without the removed item.
//
// Args:
//   - arr: The array to remove the item from.
//   - item: The item to remove.
//
// Returns:
//   - []T: The modified array without the removed item.
func Remove[T comparable](arr []T, item T) []T {
	for i, v := range arr {
		if v == item {
			return append(arr[:i], arr[i+1:]...)
		}
	}

	return arr
}

// Min is a generic function that returns the minimum value among the provided numbers.
//
// The function takes a variable number of arguments and returns the smallest value among them.
//
// Args:
//   - a: The first number to compare.
//   - b: The remaining numbers to compare.
//
// Returns:
//   - T: The minimum value among the provided numbers.
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
//
// The function takes a variable number of arguments and returns the largest value among them.
//
// Args:
//   - a: The first number to compare.
//   - b: The remaining numbers to compare.
//
// Returns:
//   - T: The maximum value among the provided numbers.
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
//
// The function attempts to parse the string as an integer.
// If the parsing is successful, it returns true, otherwise false.
//
// Args:
//   - strnum: The string to check.
//
// Returns:
//   - bool: True if the string represents a digit, otherwise false.
func IsDigit(strnum string) bool {
	_, err := strconv.ParseInt(strnum, 10, 64)
	return err == nil
}

// ParseTime parses the provided string representation of time into a [time.Time] value.
//
// The function splits the string into seconds and nanoseconds parts,
// pads the nanoseconds part with zeros to ensure 6 digits,
// and then converts the values into int64 before creating a [time.Time] object using [time.Unix].
//
// Args:
//   - strtime: The string representation of time.
//
// Returns:
//   - time.Time: The parsed time.Time object.
//   - error: An error if the parsing fails.
func ParseTime(strtime string) (t time.Time, err error) {
	sec, nsec, _ := strings.Cut(strtime, ".")
	nsec = nsec + strings.Repeat("0", 6-len(nsec))

	var secInt, nsecInt int64
	if secInt, err = strconv.ParseInt(sec, 10, 64); err != nil {
		return
	}
	if nsecInt, err = strconv.ParseInt(nsec, 10, 64); err != nil {
		return
	}
	t = time.Unix(secInt, nsecInt)

	return
}

// BoolZeroOrOne returns either "0" or "1"
//
// The function returns "0" if the provided boolean value is false, and "1" if it is true.
//
// Args:
//   - b: The boolean value to convert.
//
// Returns:
//   - string: "0" if the boolean value is false, and "1" if it is true.
func BoolZeroOrOne(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// FileExists checks whether the specified file exists.
//
// Args:
//   - filename: The name of the file to check.
//
// Returns:
//   - bool: True if the file exists, otherwise false.
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// SplitTextIntoChunks splits the provided text into chunks of the specified size.
//
// The function splits the text into chunks of the specified size, ensuring that words are not split across chunks.
// It returns a slice of strings representing the chunks.
//
// Args:
//   - text: The text to split into chunks.
//   - chunkSize: The maximum size of each chunk.
//
// Returns:
//   - []string: A slice of strings representing the chunks.
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
//
// The function calculates the added and removed flags by comparing the old and new flag values.
// It returns two integers representing the added and removed flags.
//
// Args:
//   - oldFlags: The old flag value.
//   - newFlags: The new flag value.
//
// Returns:
//   - int64: The added flags.
//   - int64: The removed flags.
func ComputeFlagChanges(oldFlags, newFlags int64) (addedFlags, removedFlags int64) {
	addedFlags = newFlags &^ oldFlags
	removedFlags = oldFlags &^ newFlags

	return
}

// LoadConfig loads the configuration from the specified file.
//
// The function reads the configuration from the specified file and unmarshals it into a Config struct.
// It returns a pointer to the Config struct and any error encountered during loading.
//
// Args:
//   - filename: The name of the configuration file.
//
// Returns:
//   - *Config: A pointer to the Config struct.
//   - error: An error if the loading fails.
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
//
// The function marshals the provided Config struct into JSON format and writes it to the specified file.
// It returns any error encountered during saving.
//
// Args:
//   - filename: The name of the configuration file.
//   - config: The Config struct to save.
//
// Returns:
//   - error: An error if the saving fails.
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

// UsernameToURL generates a URL for a Chatango resource based on the provided URL template and username.
//
// The function takes a URL template and a username as input and replaces placeholders in the template with the username's parts.
// It returns the generated URL.
//
// Args:
//   - url: The URL template with placeholders.
//   - username: The username to use for replacing placeholders.
//
// Returns:
//   - string: The generated URL.
func UsernameToURL(url, username string) string {
	path0 := username[0:1]
	path1 := username[0:1]
	if len(username) > 1 {
		path1 = username[1:2]
	}
	return fmt.Sprintf(url, path0, path1, username)
}
