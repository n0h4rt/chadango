package chadango

import (
	"context"
	"encoding/gob"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

type Persistence interface {
	Initialize() error
	Runner(context.Context)
	Close() error
	GetBotData() *SyncMap[string, any]
	GetChatData(string) *SyncMap[string, any]
	DelChatData(string)
}

// GobPersistence is responsible for loading and saving data to a gob file periodically.
// If the filename is set to an empty string, it will disable auto-saving.
// If the interval is set to less than 1 minute, it will be adjusted to 30 minutes.
type GobPersistence struct {
	Filename  string                                 // File name for the data.
	Interval  time.Duration                          // Interval for auto-saving.
	BotData   SyncMap[string, any]                   // Map to store bot-related data.
	ChatData  SyncMap[string, *SyncMap[string, any]] // Map to store chat-related data.
	context   context.Context                        // Context for running the auto save operations.
	cancelCtx context.CancelFunc                     // Function for stopping auto save operations.
}

// Load loads the data from the file into the GobPersistence struct.
func (p *GobPersistence) Load() error {
	if p.Filename == "" {
		return nil
	}

	file, err := os.Open(p.Filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)

	err = decoder.Decode(&p.BotData)
	if err != nil {
		log.Error().Str("Name", p.Filename).Err(err).Msg("GobPersistence decode BotData error.")
	}

	err = decoder.Decode(&p.ChatData)
	if err != nil {
		log.Error().Str("Name", p.Filename).Err(err).Msg("GobPersistence decode ChatData error.")
	}

	return nil
}

// Save saves the data from the GobPersistence struct to the file.
func (p *GobPersistence) Save() error {
	if p.Filename == "" {
		return nil
	}

	file, err := os.Create(p.Filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)

	err = encoder.Encode(&p.BotData)
	if err != nil {
		log.Error().Str("Name", p.Filename).Err(err).Msg("GobPersistence encode BotData error.")
	}

	err = encoder.Encode(&p.ChatData)
	if err != nil {
		log.Error().Str("Name", p.Filename).Err(err).Msg("GobPersistence encode ChatData error.")
	}

	return nil
}

// Initialize initializes the GobPersistence struct by loading the data from the file and starting the auto save routine.
func (p *GobPersistence) Initialize() error {
	p.BotData = NewSyncMap[string, any]()
	p.ChatData = NewSyncMap[string, *SyncMap[string, any]]()

	return p.Load()
}

func (p *GobPersistence) Runner(ctx context.Context) {
	if p.Filename == "" {
		return
	}

	p.context, p.cancelCtx = context.WithCancel(ctx)

	if p.Interval.Minutes() <= 0 {
		p.Interval = 30 * time.Minute
	}

	go p.autoSave()
}

// Close stops the auto save routine and saves the data to the file.
func (p *GobPersistence) Close() error {
	if p.cancelCtx != nil {
		p.cancelCtx()
	}

	return p.Save()
}

// autoSave is a goroutine that periodically saves the data to the file.
func (p *GobPersistence) autoSave() {
	ticker := time.NewTicker(p.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			log.Debug().Str("Name", p.Filename).Msg("GobPersistence auto save.")
			p.Save()
		case <-p.context.Done():
			return
		}
	}
}

// GetBotData returns a pointer to the BotData.
func (p *GobPersistence) GetBotData() *SyncMap[string, any] {
	return &p.BotData
}

// GetChatData returns a pointer to the ChatData for the given key.
// If the ChatData does not exist, it creates a new one and returns it.
func (p *GobPersistence) GetChatData(key string) *SyncMap[string, any] {
	chatData, ok := p.ChatData.Get(key)
	if !ok {
		chatData = &SyncMap[string, any]{M: map[string]any{}}
		p.ChatData.Set(key, chatData)
	}

	return chatData
}

// DelChatData deletes the ChatData for the given key.
func (p *GobPersistence) DelChatData(key string) {
	p.ChatData.Del(key)
}
