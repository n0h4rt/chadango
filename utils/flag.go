package utils

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
