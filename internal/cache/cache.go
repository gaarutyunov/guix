// Package cache provides incremental compilation caching
package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Cache stores file hashes for incremental compilation
type Cache struct {
	Hashes map[string]string `json:"hashes"`
	path   string
}

// New creates a new cache
func New(cachePath string) *Cache {
	return &Cache{
		Hashes: make(map[string]string),
		path:   cachePath,
	}
}

// Load loads the cache from disk
func Load(cachePath string) (*Cache, error) {
	c := New(cachePath)

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil // Empty cache is fine
		}
		return nil, fmt.Errorf("failed to read cache: %w", err)
	}

	if err := json.Unmarshal(data, &c.Hashes); err != nil {
		return nil, fmt.Errorf("failed to parse cache: %w", err)
	}

	return c, nil
}

// Save saves the cache to disk
func (c *Cache) Save() error {
	// Create cache directory if it doesn't exist
	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	data, err := json.MarshalIndent(c.Hashes, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	if err := os.WriteFile(c.path, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}

	return nil
}

// NeedsRegeneration checks if a file needs to be regenerated
func (c *Cache) NeedsRegeneration(srcPath string) (bool, error) {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return true, err
	}

	hash := sha256.Sum256(data)
	currentHash := hex.EncodeToString(hash[:])

	cached, exists := c.Hashes[srcPath]
	if !exists || cached != currentHash {
		c.Hashes[srcPath] = currentHash
		return true, nil
	}

	return false, nil
}

// UpdateHash updates the hash for a file
func (c *Cache) UpdateHash(srcPath string) error {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	hash := sha256.Sum256(data)
	c.Hashes[srcPath] = hex.EncodeToString(hash[:])
	return nil
}

// Remove removes a file from the cache
func (c *Cache) Remove(srcPath string) {
	delete(c.Hashes, srcPath)
}

// Clear clears all entries from the cache
func (c *Cache) Clear() {
	c.Hashes = make(map[string]string)
}
