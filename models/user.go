package models

// User represents a user object.
type User struct {
	Name   string // Name of the user.
	IsSelf bool   // Indicates if the user is self.
	IsAnon bool   // Indicates if the user is anonymous.
}
