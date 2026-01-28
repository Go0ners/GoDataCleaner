// Package config provides configuration management for GoDataCleaner.
// It loads configuration from a YAML file and/or environment variables.
// Environment variables take precedence over the config file.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
)

// Default configuration values
const (
	DefaultConfigPath            = "./config.json"
	DefaultLocalHost             = "localhost"
	DefaultLocalPort             = 61913
	DefaultQBittorrentHost       = "qbt.home"
	DefaultQBittorrentPort       = 80
	DefaultQBittorrentUsername   = "admin"
	DefaultQBittorrentPassword   = "adminadmin"
	DefaultQBittorrentMaxWorkers = 10
	DefaultSQLitePath            = "./data/torrents.db"
	DefaultSQLiteBatchSize       = 1000
	DefaultLocalPath             = "./data/torrents"
)

// Error definitions for configuration validation
var (
	ErrInvalidPort = errors.New("invalid port: must be between 1 and 65535")
	ErrInvalidPath = errors.New("invalid path: path cannot be empty")
)

// Config holds the application configuration.
type Config struct {
	LocalHost             string `json:"local_host"`
	LocalPort             int    `json:"local_port"`
	QBittorrentHost       string `json:"qbittorrent_host"`
	QBittorrentPort       int    `json:"qbittorrent_port"`
	QBittorrentUsername   string `json:"qbittorrent_username"`
	QBittorrentPassword   string `json:"qbittorrent_password"`
	QBittorrentMaxWorkers int    `json:"qbittorrent_max_workers"`
	SQLitePath            string `json:"sqlite_path"`
	SQLiteBatchSize       int    `json:"sqlite_batch_size"`
	LocalPath             string `json:"local_path"`
}

// Load loads the configuration with the following priority:
// 1. Environment variables (highest priority)
// 2. Config file (config.json)
// 3. Default values (lowest priority)
func Load() (*Config, error) {
	// Start with defaults
	cfg := &Config{
		LocalHost:             DefaultLocalHost,
		LocalPort:             DefaultLocalPort,
		QBittorrentHost:       DefaultQBittorrentHost,
		QBittorrentPort:       DefaultQBittorrentPort,
		QBittorrentUsername:   DefaultQBittorrentUsername,
		QBittorrentPassword:   DefaultQBittorrentPassword,
		QBittorrentMaxWorkers: DefaultQBittorrentMaxWorkers,
		SQLitePath:            DefaultSQLitePath,
		SQLiteBatchSize:       DefaultSQLiteBatchSize,
		LocalPath:             DefaultLocalPath,
	}

	// Load from config file if it exists
	configPath := getEnvString("CONFIG_PATH", DefaultConfigPath)
	if err := cfg.loadFromFile(configPath); err != nil {
		// Ignore file not found errors
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Override with environment variables (highest priority)
	cfg.loadFromEnv()

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// loadFromFile loads configuration from a JSON file.
func (c *Config) loadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Parse JSON into a temporary struct to preserve zero values
	var fileCfg Config
	if err := json.Unmarshal(data, &fileCfg); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Only override non-zero values from file
	if fileCfg.LocalHost != "" {
		c.LocalHost = fileCfg.LocalHost
	}
	if fileCfg.LocalPort != 0 {
		c.LocalPort = fileCfg.LocalPort
	}
	if fileCfg.QBittorrentHost != "" {
		c.QBittorrentHost = fileCfg.QBittorrentHost
	}
	if fileCfg.QBittorrentPort != 0 {
		c.QBittorrentPort = fileCfg.QBittorrentPort
	}
	if fileCfg.QBittorrentUsername != "" {
		c.QBittorrentUsername = fileCfg.QBittorrentUsername
	}
	if fileCfg.QBittorrentPassword != "" {
		c.QBittorrentPassword = fileCfg.QBittorrentPassword
	}
	if fileCfg.QBittorrentMaxWorkers != 0 {
		c.QBittorrentMaxWorkers = fileCfg.QBittorrentMaxWorkers
	}
	if fileCfg.SQLitePath != "" {
		c.SQLitePath = fileCfg.SQLitePath
	}
	if fileCfg.SQLiteBatchSize != 0 {
		c.SQLiteBatchSize = fileCfg.SQLiteBatchSize
	}
	if fileCfg.LocalPath != "" {
		c.LocalPath = fileCfg.LocalPath
	}

	return nil
}

// loadFromEnv overrides configuration with environment variables.
func (c *Config) loadFromEnv() {
	if v := os.Getenv("LOCAL_HOST"); v != "" {
		c.LocalHost = v
	}
	if v := os.Getenv("LOCAL_PORT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			c.LocalPort = i
		}
	}
	if v := os.Getenv("QBITTORRENT_HOST"); v != "" {
		c.QBittorrentHost = v
	}
	if v := os.Getenv("QBITTORRENT_PORT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			c.QBittorrentPort = i
		}
	}
	if v := os.Getenv("QBITTORRENT_USERNAME"); v != "" {
		c.QBittorrentUsername = v
	}
	if v := os.Getenv("QBITTORRENT_PASSWORD"); v != "" {
		c.QBittorrentPassword = v
	}
	if v := os.Getenv("QBITTORRENT_MAX_WORKERS"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			c.QBittorrentMaxWorkers = i
		}
	}
	if v := os.Getenv("SQLITE_PATH"); v != "" {
		c.SQLitePath = v
	}
	if v := os.Getenv("SQLITE_BATCH_SIZE"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			c.SQLiteBatchSize = i
		}
	}
	if v := os.Getenv("LOCAL_PATH"); v != "" {
		c.LocalPath = v
	}
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if !isValidPort(c.LocalPort) {
		return fmt.Errorf("LOCAL_PORT %w: got %d", ErrInvalidPort, c.LocalPort)
	}
	if !isValidPort(c.QBittorrentPort) {
		return fmt.Errorf("QBITTORRENT_PORT %w: got %d", ErrInvalidPort, c.QBittorrentPort)
	}
	if c.SQLitePath == "" {
		return fmt.Errorf("SQLITE_PATH %w", ErrInvalidPath)
	}
	if c.LocalPath == "" {
		return fmt.Errorf("LOCAL_PATH %w", ErrInvalidPath)
	}
	if c.QBittorrentMaxWorkers < 1 {
		return fmt.Errorf("QBITTORRENT_MAX_WORKERS must be at least 1: got %d", c.QBittorrentMaxWorkers)
	}
	if c.SQLiteBatchSize < 1 {
		return fmt.Errorf("SQLITE_BATCH_SIZE must be at least 1: got %d", c.SQLiteBatchSize)
	}
	return nil
}

// QBittorrentURL returns the full qBittorrent server URL.
func (c *Config) QBittorrentURL() string {
	// Don't include port 80 explicitly as it can cause auth issues with some servers
	if c.QBittorrentPort == 80 {
		return fmt.Sprintf("http://%s", c.QBittorrentHost)
	}
	if c.QBittorrentPort == 443 {
		return fmt.Sprintf("https://%s", c.QBittorrentHost)
	}
	return fmt.Sprintf("http://%s:%d", c.QBittorrentHost, c.QBittorrentPort)
}

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func isValidPort(port int) bool {
	return port >= 1 && port <= 65535
}
