package utils

// Number represents a numeric type.
type Number interface {
	int | int16 | int32 | int64 | float32 | float64
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
