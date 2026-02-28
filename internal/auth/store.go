package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const tokenFileName = "token.json"

// TokenData holds the OAuth token and related info.
type TokenData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	WorkspaceID  string `json:"workspace_id,omitempty"`
	BotID        string `json:"bot_id,omitempty"`
}

// configDir returns ~/.config/noti
func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".config", "noti"), nil
}

// tokenPath returns the full path to the token file.
func tokenPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, tokenFileName), nil
}

// SaveToken persists the OAuth token to disk.
func SaveToken(data *TokenData) error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	path := filepath.Join(dir, tokenFileName)
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}
	if err := os.WriteFile(path, b, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}
	return nil
}

// LoadToken reads the stored OAuth token from disk.
// Returns nil if no token file exists.
func LoadToken() (*TokenData, error) {
	path, err := tokenPath()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}
	var data TokenData
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, fmt.Errorf("failed to parse token file: %w", err)
	}
	return &data, nil
}

// DeleteToken removes the stored token file.
func DeleteToken() error {
	path, err := tokenPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete token file: %w", err)
	}
	return nil
}
