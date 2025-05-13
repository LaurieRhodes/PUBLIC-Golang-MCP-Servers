package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	AllowedDirectories []string `json:"allowedDirectories"`
}

// Default config file name
const configFileName = "config.json"

// ErrNoAllowedDirectories is returned when no allowed directories are specified
var ErrNoAllowedDirectories = errors.New("at least one allowed directory must be specified in config.json")

// LoadConfig loads the configuration from a JSON file in the executable directory
func LoadConfig() (*Config, error) {
	// Get the directory of the executable
	executablePath, err := getExecutablePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}

	// Print the executable path for debugging
	fmt.Fprintf(os.Stderr, "Executable directory: %s\n", executablePath)

	// Build the path to the config file
	configFilePath := filepath.Join(executablePath, configFileName)
	fmt.Fprintf(os.Stderr, "Looking for config file at: %s\n", configFilePath)

	// Check if the config file exists
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		// Try in current working directory as fallback
		cwd, err := os.Getwd()
		if err == nil {
			cwdConfigPath := filepath.Join(cwd, configFileName)
			fmt.Fprintf(os.Stderr, "Config not found in executable directory, checking current directory: %s\n", cwdConfigPath)
			
			if _, err := os.Stat(cwdConfigPath); err == nil {
				// Found config in current directory
				configFilePath = cwdConfigPath
				fmt.Fprintf(os.Stderr, "Found config file in current directory\n")
			} else {
				// Create a default config if none exists
				fmt.Fprintf(os.Stderr, "No config file found, creating default in executable directory\n")
				return createDefaultConfig(configFilePath)
			}
		} else {
			// Couldn't get current directory, create config in executable directory
			fmt.Fprintf(os.Stderr, "No config file found, creating default in executable directory\n")
			return createDefaultConfig(configFilePath)
		}
	}

	// Read the config file
	fmt.Fprintf(os.Stderr, "Reading config from: %s\n", configFilePath)
	file, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the config file
	config := &Config{}
	if err := json.Unmarshal(file, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate the config
	if len(config.AllowedDirectories) == 0 {
		return nil, ErrNoAllowedDirectories
	}

	// Resolve and validate all directory paths
	resolvedDirs := make([]string, 0, len(config.AllowedDirectories))
	for _, dir := range config.AllowedDirectories {
		// Convert to absolute path
		absPath, err := filepath.Abs(dir)
		if err != nil {
			return nil, fmt.Errorf("error resolving path %s: %w", dir, err)
		}

		// Check if it exists and is a directory
		info, err := os.Stat(absPath)
		if err != nil {
			return nil, fmt.Errorf("error accessing directory %s: %w", absPath, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("error: %s is not a directory", absPath)
		}

		resolvedDirs = append(resolvedDirs, absPath)
	}
	
	// Update the config with resolved paths
	config.AllowedDirectories = resolvedDirs

	fmt.Fprintf(os.Stderr, "Configuration loaded successfully\n")
	return config, nil
}

// createDefaultConfig creates a default config file with example allowed directories
func createDefaultConfig(configFilePath string) (*Config, error) {
	// Get current directory as an example
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "C:\\path\\to\\allowed\\directory"
	}
	
	config := &Config{
		AllowedDirectories: []string{cwd},
	}

	// Convert config to JSON
	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal default config: %w", err)
	}

	// Write the config file
	if err := os.WriteFile(configFilePath, jsonData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write default config file: %w", err)
	}

	// Return the config with an error message to edit it
	return nil, fmt.Errorf("created default config file at %s. Please edit this file to add your allowed directories", configFilePath)
}

// getExecutablePath returns the directory of the current executable
func getExecutablePath() (string, error) {
	// Get the absolute path to the executable
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	
	// Resolve any symbolic links
	realPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		// If we can't resolve symlinks, use the original path
		realPath = execPath
	}
	
	// Get the directory containing the executable
	return filepath.Dir(realPath), nil
}
