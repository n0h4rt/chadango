package chadango

// EventType represents the type of an event.
type EventType int64

// Event types.
const (
	// Event triggered when the application starts.
	OnStart EventType = 1 << iota
	// Event triggered when the application stops.
	OnStop
	// Event triggered when an error occurs.
	OnError

	// Event triggered when the bot joins a group.
	OnGroupJoined
	// Event triggered when the bot leaves a group.
	OnGroupLeft
	// Event triggered when the bot reconnects to a group.
	OnGroupReconnected
	// Event triggered when a user joins a group.
	OnJoin
	// Event triggered when a user logs in.
	OnLogin
	// Event triggered when a user logs out.
	OnLogout
	// Event triggered when a user leaves a group.
	OnLeave
	// Event triggered when the participant count in a group changes.
	OnParticipantCountChange
	// Event triggered when a message is received.
	OnMessage
	// Event triggered when a message is deleted.
	OnMessageDelete
	// Event triggered when message history is retrieved.
	OnMessageHistory
	// Event triggered when a message is updated.
	OnMessageUpdate
	// Event triggered when an announcement is received.
	OnAnnouncement
	// Event triggered when the group information is updated.
	OnUpdateGroupInfo
	// Event triggered when flags are updated.
	OnFlagUpdate
	// Event triggered when a moderator is added.
	OnModeratorAdded
	// Event triggered when a moderator is updated.
	OnModeratorUpdated
	// Event triggered when a moderator is removed.
	OnModeratorRemoved
	// Event triggered when all messages are cleared.
	OnClearAll
	// Event triggered when a user is banned.
	OnUserBanned
	// Event triggered when a user is unbanned.
	OnUserUnbanned
	// Event triggered when all users are unbanned.
	OnAllUserUnbanned

	// Event triggered when the bot connects to a private chat.
	OnPrivateConnected
	// Event triggered when the bot disconnects from a private chat.
	OnPrivateDisconnected
	// Event triggered when the bot reconnects to a private chat.
	OnPrivateReconnected
	// Event triggered when the bot is kicked off from a private chat.
	OnPrivateKickedOff
	// Event triggered when a private message is received.
	OnPrivateMessage
	// Event triggered when an offline private message is received.
	OnPrivateOfflineMessage
	// Event triggered when a friend becomes online in a private chat.
	OnPrivateFriendOnline
	// Event triggered when a friend becomes online in a private chat app.
	OnPrivateFriendOnlineApp
	// Event triggered when a friend goes offline in a private chat.
	OnPrivateFriendOffline
	// Event triggered when a friend becomes active in a private chat.
	OnPrivateFriendActive
	// Event triggered when a friend becomes idle in a private chat.
	OnPrivateFriendIdle

	// Event triggered when the user profile is updated.
	OnUpdateUserProfile
)

// String returns a string of said EventType.
func (e EventType) String() string {
	switch e {
	case OnStart:
		return "OnStart"
	case OnStop:
		return "OnStop"
	case OnError:
		return "OnError"
	case OnGroupJoined:
		return "OnGroupJoined"
	case OnGroupLeft:
		return "OnGroupLeft"
	case OnGroupReconnected:
		return "OnGroupReconnected"
	case OnJoin:
		return "OnJoin"
	case OnLogin:
		return "OnLogin"
	case OnLogout:
		return "OnLogout"
	case OnLeave:
		return "OnLeave"
	case OnParticipantCountChange:
		return "OnParticipantCountChange"
	case OnMessage:
		return "OnMessage"
	case OnMessageDelete:
		return "OnMessageDelete"
	case OnMessageHistory:
		return "OnMessageHistory"
	case OnMessageUpdate:
		return "OnMessageUpdate"
	case OnAnnouncement:
		return "OnAnnouncement"
	case OnUpdateGroupInfo:
		return "OnUpdateGroupInfo"
	case OnFlagUpdate:
		return "OnFlagUpdate"
	case OnModeratorAdded:
		return "OnModeratorAdded"
	case OnModeratorUpdated:
		return "OnModeratorUpdated"
	case OnModeratorRemoved:
		return "OnModeratorRemoved"
	case OnClearAll:
		return "OnClearAll"
	case OnUserBanned:
		return "OnUserBanned"
	case OnUserUnbanned:
		return "OnUserUnbanned"
	case OnAllUserUnbanned:
		return "OnAllUserUnbanned"
	case OnPrivateConnected:
		return "OnPrivateConnected"
	case OnPrivateDisconnected:
		return "OnPrivateDisconnected"
	case OnPrivateReconnected:
		return "OnPrivateReconnected"
	case OnPrivateKickedOff:
		return "OnPrivateKickedOff"
	case OnPrivateMessage:
		return "OnPrivateMessage"
	case OnPrivateOfflineMessage:
		return "OnPrivateOfflineMessage"
	case OnPrivateFriendOnline:
		return "OnPrivateFriendOnline"
	case OnPrivateFriendOnlineApp:
		return "OnPrivateFriendOnlineApp"
	case OnPrivateFriendOffline:
		return "OnPrivateFriendOffline"
	case OnPrivateFriendActive:
		return "OnPrivateFriendActive"
	case OnPrivateFriendIdle:
		return "OnPrivateFriendIdle"
	case OnUpdateUserProfile:
		return "OnUpdateUserProfile"
	default:
		return "UnknownEvent"
	}
}

// Event represents an event that can occur in the application.
type Event struct {
	Type             EventType    // The type of the event.
	IsPrivate        bool         // Indicates if the event is related to a private chat.
	Private          *Private     // The private chat associated with the event.
	Group            *Group       // The group associated with the event.
	User             *User        // The user associated with the event.
	Message          *Message     // The message associated with the event.
	Command          string       // The command associated with the event.
	WithArgument     bool         // Indicates if the command has an argument.
	Argument         string       // The argument associated with the command.
	Arguments        []string     // The arguments associated with the command.
	Participant      *Participant // The participant associated with the event.
	FlagAdded        int64        // The flags added in the event.
	FlagRemoved      int64        // The flags removed in the event.
	Blocked          *Blocked     // The blocked user associated with the event.
	Unblocked        *Unblocked   // The unblocked user associated with the event.
	GroupInfo        *GroupInfo   // The group information associated with the event.
	ModGrantedAccess int64        // The granted moderator access level associated with the event.
	ModRevokedAccess int64        // The revoked moderator access level associated with the event.
	Error            any          // The error associated with the event.
}
