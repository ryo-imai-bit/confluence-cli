package client

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the Confluence API configuration
type Config struct {
	BaseURL  string `yaml:"base_url"`
	Email    string `yaml:"email"`
	APIToken string `yaml:"api_token"`
}

// Client is an HTTP client for Confluence API
type Client struct {
	httpClient *http.Client
	config     Config
}

// NewClient creates a new Confluence API client.
// If the base URL is a *.atlassian.net URL, it automatically resolves the Cloud ID
// and rewrites the URL to the api.atlassian.com gateway for scoped API token support.
func NewClient() (*Client, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	resolvedURL, err := ResolveBaseURL(httpClient, config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve Confluence API URL: %w", err)
	}
	config.BaseURL = resolvedURL

	return &Client{
		httpClient: httpClient,
		config:     config,
	}, nil
}

// LocalConfigName is the name of the project-local config file
const LocalConfigName = ".confluence-cli.yaml"

// ConfigPath returns the path to the user config file
func ConfigPath() string {
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "confluence-cli", "config.yaml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "confluence-cli", "config.yaml")
}

// FindLocalConfig searches for a local config file in current and parent directories
func FindLocalConfig() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		configPath := filepath.Join(dir, LocalConfigName)
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// LoadConfig loads configuration with the following precedence (highest first):
// 1. Environment variables
// 2. User config file (~/.config/confluence-cli/config.yaml)
// 3. Project-local config file (.confluence-cli.yaml)
func LoadConfig() (Config, error) {
	config := Config{}

	// Load project-local config first (lowest priority for files)
	if localPath := FindLocalConfig(); localPath != "" {
		if data, err := os.ReadFile(localPath); err == nil {
			yaml.Unmarshal(data, &config)
		}
	}

	// Override with user config file
	userConfigPath := ConfigPath()
	if data, err := os.ReadFile(userConfigPath); err == nil {
		var userConfig Config
		if err := yaml.Unmarshal(data, &userConfig); err != nil {
			return config, fmt.Errorf("failed to parse user config file: %w", err)
		}
		// Merge: user config overrides local config
		if userConfig.BaseURL != "" {
			config.BaseURL = userConfig.BaseURL
		}
		if userConfig.Email != "" {
			config.Email = userConfig.Email
		}
		if userConfig.APIToken != "" {
			config.APIToken = userConfig.APIToken
		}
	}

	// Override with environment variables (highest priority)
	if v := os.Getenv("CONFLUENCE_BASE_URL"); v != "" {
		config.BaseURL = v
	}
	if v := os.Getenv("CONFLUENCE_EMAIL"); v != "" {
		config.Email = v
	}
	if v := os.Getenv("CONFLUENCE_API_TOKEN"); v != "" {
		config.APIToken = v
	}

	// Validate required fields
	if config.BaseURL == "" {
		return config, errors.New("base_url is required (set in config file or CONFLUENCE_BASE_URL env var)")
	}
	if config.Email == "" {
		return config, errors.New("email is required (set in config file or CONFLUENCE_EMAIL env var)")
	}
	if config.APIToken == "" {
		return config, errors.New("api_token is required (set in config file or CONFLUENCE_API_TOKEN env var)")
	}

	// Remove trailing slash from base URL
	config.BaseURL = strings.TrimSuffix(config.BaseURL, "/")

	return config, nil
}

// SaveConfig saves the configuration to the config file
func SaveConfig(config Config) error {
	configPath := ConfigPath()

	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write with restrictive permissions (owner read/write only)
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ValidateConfig checks if configuration is valid
func ValidateConfig() error {
	_, err := LoadConfig()
	return err
}

// basicAuthHeader generates a Basic Auth header value
func (c *Client) basicAuthHeader() string {
	auth := c.config.Email + ":" + c.config.APIToken
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

// Do performs an HTTP request with authentication
func (c *Client) Do(method, path string, body io.Reader) (*http.Response, error) {
	url := c.config.BaseURL + path

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", c.basicAuthHeader())
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// Get performs a GET request
func (c *Client) Get(path string) (*http.Response, error) {
	return c.Do(http.MethodGet, path, nil)
}

// Post performs a POST request
func (c *Client) Post(path string, body io.Reader) (*http.Response, error) {
	return c.Do(http.MethodPost, path, body)
}

// Put performs a PUT request
func (c *Client) Put(path string, body io.Reader) (*http.Response, error) {
	return c.Do(http.MethodPut, path, body)
}

// Delete performs a DELETE request
func (c *Client) Delete(path string) (*http.Response, error) {
	return c.Do(http.MethodDelete, path, nil)
}
