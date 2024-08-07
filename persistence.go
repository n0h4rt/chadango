package chadango

import (
	"context"
	"encoding/gob"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

// Persistence is an interface that defines the methods for managing data persistence.
//
// It provides methods for initializing, running, closing, and accessing data stored in the persistence layer.
type Persistence interface {
	Initialize() error                        // Initialize initializes the persistence layer.
	Runner(context.Context)                   // Runner starts a goroutine that manages the persistence layer.
	Close() error                             // Close closes the persistence layer and saves any pending data.
	GetBotData() *SyncMap[string, any]        // GetBotData returns a pointer to the bot-related data.
	GetChatData(string) *SyncMap[string, any] // GetChatData returns a pointer to the chat-related data for the given key.
	DelChatData(string)                       // DelChatData deletes the chat-related data for the given key.
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
//
// Args:
//   - none
//
// Returns:
//   - error: An error if loading fails.
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
//
// Returns:
//   - error: An error if saving fails.
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
//
// Returns:
//   - error: An error if initialization fails.
func (p *GobPersistence) Initialize() error {
	p.BotData = NewSyncMap[string, any]()
	p.ChatData = NewSyncMap[string, *SyncMap[string, any]]()

	return p.Load()
}

// Runner starts a goroutine that manages the persistence layer.
//
// Args:
//   - ctx: The context for running the auto save operations.
//
// Returns:
//   - none
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
//
// Returns:
//   - error: An error if closing fails.
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
//
// Returns:
//   - *SyncMap[string, any]: A pointer to the bot-related data.
func (p *GobPersistence) GetBotData() *SyncMap[string, any] {
	return &p.BotData
}

// GetChatData returns a pointer to the ChatData for the given key.
// If the ChatData does not exist, it creates a new one and returns it.
//
// Args:
//   - key: The key to retrieve the ChatData for.
//
// Returns:
//   - *SyncMap[string, any]: A pointer to the chat-related data for the given key.
func (p *GobPersistence) GetChatData(key string) *SyncMap[string, any] {
	chatData, ok := p.ChatData.Get(key)
	if !ok {
		chatData = &SyncMap[string, any]{M: map[string]any{}}
		p.ChatData.Set(key, chatData)
	}

	return chatData
}

// DelChatData deletes the ChatData for the given key.
//
// Args:
//   - key: The key to delete the ChatData for.
func (p *GobPersistence) DelChatData(key string) {
	p.ChatData.Del(key)
}
