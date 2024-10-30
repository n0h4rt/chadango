package models

import "time"

// Participant represents a group participant object.
type Participant struct {
	ParticipantID string    // Participant ID.
	UserID        int       // User ID.
	User          *User     // User object.
	Time          time.Time // Time of participation.
}
