package utils

import (
	"strconv"
	"strings"
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
