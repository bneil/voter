package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"voter/internal/models"
)

// JSONGameStore implements GameStore using JSON file storage
type JSONGameStore struct {
	dataDir string
	mu      sync.RWMutex
}

// NewJSONGameStore creates a new JSON-based game store
func NewJSONGameStore(dataDir string) (*JSONGameStore, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &JSONGameStore{
		dataDir: dataDir,
	}, nil
}

// SaveGame saves a game to storage
func (s *JSONGameStore) SaveGame(game *models.Game) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filename := filepath.Join(s.dataDir, fmt.Sprintf("game_%s.json", game.ID))

	data, err := json.MarshalIndent(game, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal game: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write game file: %w", err)
	}

	return nil
}

// GetGame retrieves a game from storage
func (s *JSONGameStore) GetGame(id string) (*models.Game, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filename := filepath.Join(s.dataDir, fmt.Sprintf("game_%s.json", id))

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("game not found: %w", err)
		}
		return nil, fmt.Errorf("failed to read game file: %w", err)
	}

	var game models.Game
	if err := json.Unmarshal(data, &game); err != nil {
		return nil, fmt.Errorf("failed to unmarshal game: %w", err)
	}

	return &game, nil
}

// ListGames returns all games
func (s *JSONGameStore) ListGames() ([]*models.Game, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	files, err := filepath.Glob(filepath.Join(s.dataDir, "game_*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list game files: %w", err)
	}

	var games []*models.Game
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue // Skip files that can't be read
		}

		var game models.Game
		if err := json.Unmarshal(data, &game); err != nil {
			continue // Skip files that can't be unmarshaled
		}

		games = append(games, &game)
	}

	return games, nil
}

// DeleteGame removes a game from storage
func (s *JSONGameStore) DeleteGame(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filename := filepath.Join(s.dataDir, fmt.Sprintf("game_%s.json", id))

	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete game file: %w", err)
	}

	return nil
}
