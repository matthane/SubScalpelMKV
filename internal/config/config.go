package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	DefaultLanguages []string           `yaml:"default_languages"`
	OutputTemplate   string             `yaml:"output_template"`
	OutputDir        string             `yaml:"output_dir"`
	Profiles         map[string]Profile `yaml:"profiles"`
}

// Profile represents a named configuration profile
type Profile struct {
	Languages      []string `yaml:"languages"`
	OutputTemplate string   `yaml:"output_template"`
	OutputDir      string   `yaml:"output_dir"`
}

// AppliedConfig represents the final configuration after merging defaults, config file, and CLI flags
type AppliedConfig struct {
	Languages      []string
	OutputTemplate string
	OutputDir      string
}

// GetDefaultConfig returns the default configuration values
func GetDefaultConfig() Config {
	return Config{
		DefaultLanguages: []string{},
		OutputTemplate:   "",
		OutputDir:        "",
		Profiles:         make(map[string]Profile),
	}
}

// FindConfigFile searches for configuration files in standard locations
func FindConfigFile() string {
	// 1. Current directory (highest priority)
	if _, err := os.Stat("./subscalpelmkv.yaml"); err == nil {
		return "./subscalpelmkv.yaml"
	}

	// 2. OS-specific config directory
	if configDir, err := os.UserConfigDir(); err == nil {
		path := filepath.Join(configDir, "subscalpelmkv", "config.yaml")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// 3. Home directory dot-file
	if homeDir, err := os.UserHomeDir(); err == nil {
		path := filepath.Join(homeDir, ".subscalpelmkv.yaml")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "" // No config found
}

// LoadConfig loads configuration from the specified file path
func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		// Return default config if no path specified
		config := GetDefaultConfig()
		return &config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Ensure Profiles map is initialized
	if config.Profiles == nil {
		config.Profiles = make(map[string]Profile)
	}

	return &config, nil
}

// LoadConfigWithFallback attempts to find and load a config file, returns default config if none found
func LoadConfigWithFallback() (*Config, error) {
	configPath := FindConfigFile()
	return LoadConfig(configPath)
}

// GetProfile returns the specified profile from the config, or an error if not found
func (c *Config) GetProfile(profileName string) (Profile, error) {
	if profile, exists := c.Profiles[profileName]; exists {
		return profile, nil
	}
	return Profile{}, fmt.Errorf("profile '%s' not found in configuration", profileName)
}

// ApplyProfile merges a profile with the base config and returns the applied configuration
func (c *Config) ApplyProfile(profileName string) (*AppliedConfig, error) {
	profile, err := c.GetProfile(profileName)
	if err != nil {
		return nil, err
	}

	applied := &AppliedConfig{
		Languages:      c.DefaultLanguages,
		OutputTemplate: c.OutputTemplate,
		OutputDir:      c.OutputDir,
	}

	// Override with profile values if they're set
	if len(profile.Languages) > 0 {
		applied.Languages = profile.Languages
	}
	if profile.OutputTemplate != "" {
		applied.OutputTemplate = profile.OutputTemplate
	}
	if profile.OutputDir != "" {
		applied.OutputDir = profile.OutputDir
	}

	return applied, nil
}

// ApplyDefaults returns the default configuration as applied config
func (c *Config) ApplyDefaults() *AppliedConfig {
	return &AppliedConfig{
		Languages:      c.DefaultLanguages,
		OutputTemplate: c.OutputTemplate,
		OutputDir:      c.OutputDir,
	}
}

// ValidateConfig performs basic validation on the configuration
func ValidateConfig(config *Config) error {
	// Validate profiles
	for profileName, profile := range config.Profiles {
		if profileName == "" {
			return fmt.Errorf("profile name cannot be empty")
		}
		
		// Validate language codes in profile
		for _, lang := range profile.Languages {
			if len(lang) != 2 && len(lang) != 3 {
				return fmt.Errorf("invalid language code '%s' in profile '%s': must be 2 or 3 characters", lang, profileName)
			}
		}
	}
	
	// Validate default language codes
	for _, lang := range config.DefaultLanguages {
		if len(lang) != 2 && len(lang) != 3 {
			return fmt.Errorf("invalid default language code '%s': must be 2 or 3 characters", lang)
		}
	}

	return nil
}

// GetConfigLocations returns all possible config file locations for display to users
func GetConfigLocations() []string {
	locations := []string{
		"./subscalpelmkv.yaml (current directory)",
	}

	if configDir, err := os.UserConfigDir(); err == nil {
		locations = append(locations, filepath.Join(configDir, "subscalpelmkv", "config.yaml"))
	}

	if homeDir, err := os.UserHomeDir(); err == nil {
		locations = append(locations, filepath.Join(homeDir, ".subscalpelmkv.yaml"))
	}

	return locations
}

// CLIFlags represents the command line flags that can be overridden by config
type CLIFlags struct {
	Languages      []string
	OutputTemplate string
	OutputDir      string
}

// MergeWithCLI merges applied configuration with CLI flags, where CLI flags take precedence
func (ac *AppliedConfig) MergeWithCLI(cli CLIFlags) *AppliedConfig {
	merged := &AppliedConfig{
		Languages:      ac.Languages,
		OutputTemplate: ac.OutputTemplate,
		OutputDir:      ac.OutputDir,
	}

	// CLI flags override config values if they're set
	if len(cli.Languages) > 0 {
		merged.Languages = cli.Languages
	}
	if cli.OutputTemplate != "" {
		merged.OutputTemplate = cli.OutputTemplate
	}
	if cli.OutputDir != "" {
		merged.OutputDir = cli.OutputDir
	}

	return merged
}