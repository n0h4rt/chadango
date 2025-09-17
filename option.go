package chadango

// Option represents a configurable parameter for the Application.
type Option func(*Application)

// WithPersistence enables the persistence layer for the application.
//
// Args:
//   - persistence: The persistence layer to use for the application.
//
// Returns:
//   - Option: A function that applies the specified persistence layer to the Application.
func WithPersistence(persistence Persistence) Option {
	return func(a *Application) {
		a.persistence = persistence
	}
}

// WithDebug enables debug mode for the application.
//
// When debug mode is enabled, the application may log additional information
// for troubleshooting and development purposes.
//
// Returns:
//   - Option: A function that enables debug mode for the Application.
func WithDebug() Option {
	return func(a *Application) {
		a.isDebug = true
	}
}

// WithEnablePM enables private messaging (PM) for the application.
//
// When private messaging is enabled, the application will process and handle
// messages sent privately to it.
//
// Returns:
//   - Option: A function that enables private messaging for the Application.
func WithEnablePM() Option {
	return func(a *Application) {
		a.enablePM = true
	}
}

// WithPMSessionID sets a specific session ID for private messaging.
//
// Args:
//   - sessionID: A string representing the session ID to be used for
//                private messaging sessions.
//
// Returns:
//   - Option: A function that sets the private messaging session ID
//             for the Application.
func WithPMSessionID(sessionID string) Option {
	return func(a *Application) {
		a.pmSessionID = sessionID
	}
}
