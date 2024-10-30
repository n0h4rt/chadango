package models

import "time"

// UserStatus represents the online status of a user.
type UserStatus struct {
	User *User         // User object.
	Time time.Time     // Time of the status.
	Info string        // Additional information about the status.
	Idle time.Duration // Idle duration of the user.
}
