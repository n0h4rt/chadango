package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

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
