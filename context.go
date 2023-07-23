package chadango

// Context represents the context for event callback.
type Context struct {
	App      *Application          // App is a pointer to the Application that manages this context.
	ChatData *SyncMap[string, any] // ChatData is a synchronized map to store data specific to the chat.
	BotData  *SyncMap[string, any] // BotData is a synchronized map to store data specific to the bot.
}
