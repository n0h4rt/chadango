package chadango

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents a configuration object.
type Config struct {
	Username  string   `json:"username"`  // Username of the configuration.
	Password  string   `json:"password"`  // Password of the configuration.
	AnonName  string   `json:"anonname"`  // Anonymous name of the configuration.
	Groups    []string `json:"groups"`    // List of groups in the configuration.
	NameColor string   `json:"namecolor"` // Name color in the configuration.
	TextColor string   `json:"textcolor"` // Text color in the configuration.
	TextFont  string   `json:"textfont"`  // Text font in the configuration.
	TextSize  int      `json:"textsize"`  // Text size in the configuration.
	SessionID string   `json:"sessionid"` // Session ID in the configuration.
	EnableBG  bool     `json:"enablebg"`  // Enable background in the configuration.
	EnablePM  bool     `json:"enablepm"`  // Enable private messages in the configuration.
	Debug     bool     `json:"debug"`     // Debug mode in the configuration.
	Prefix    string   `json:"prefix"`    // Prefix for commands in the configuration.
}

// LoadConfig loads the configuration from the specified file.
//
// The function reads the configuration from the specified file and unmarshals it into a Config struct.
// It returns a pointer to the Config struct and any error encountered during loading.
//
// Args:
//   - filename: The name of the configuration file.
//
// Returns:
//   - *Config: A pointer to the Config struct.
//   - error: An error if the loading fails.
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config data: %v", err)
	}

	return &config, nil
}

// SaveConfig saves the configuration to the specified file.
//
// The function marshals the provided Config struct into JSON format and writes it to the specified file.
// It returns any error encountered during saving.
//
// Args:
//   - filename: The name of the configuration file.
//   - config: The Config struct to save.
//
// Returns:
//   - error: An error if the saving fails.
func SaveConfig(filename string, config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config data: %v", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}
