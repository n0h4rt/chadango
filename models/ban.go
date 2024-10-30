package models

import "time"

// Blocked represents a banned user in a group.
type Blocked struct {
	IP           string    // IP address of the blocked user.
	ModerationID string    // Moderation ID of the block.
	Target       string    // Target username.
	Blocker      string    // Username of the blocker.
	Time         time.Time // Time of blocking.
}

// Unblocked represents an unblocked user in a group.
type Unblocked struct {
	IP           string    // IP address of the unblocked user.
	ModerationID string    // Moderation ID of the unblock.
	Target       string    // Target username.
	Unblocker    string    // Username of the unblocker.
	Time         time.Time // Time of unblocking.
}
