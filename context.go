package chadango

type Context struct {
	App      *Application
	ChatData *SyncMap[string, any]
	BotData  *SyncMap[string, any]
}
