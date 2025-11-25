package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"voter/internal/models"
)

// JSONProjectStore implements ProjectStore using JSON file storage
type JSONProjectStore struct {
	dataDir string
	mu      sync.RWMutex
}

// NewJSONProjectStore creates a new JSON-based project store
func NewJSONProjectStore(dataDir string) (*JSONProjectStore, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &JSONProjectStore{
		dataDir: dataDir,
	}, nil
}

// SaveProject saves a project to storage
func (s *JSONProjectStore) SaveProject(project *models.Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filename := filepath.Join(s.dataDir, fmt.Sprintf("project_%s.json", project.ID))

	data, err := json.MarshalIndent(project, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal project: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write project file: %w", err)
	}

	return nil
}

// GetProject retrieves a project from storage
func (s *JSONProjectStore) GetProject(id string) (*models.Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filename := filepath.Join(s.dataDir, fmt.Sprintf("project_%s.json", id))

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("project not found: %w", err)
		}
		return nil, fmt.Errorf("failed to read project file: %w", err)
	}

	var project models.Project
	if err := json.Unmarshal(data, &project); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project: %w", err)
	}

	return &project, nil
}

// ListProjects returns all projects
func (s *JSONProjectStore) ListProjects() ([]*models.Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	files, err := filepath.Glob(filepath.Join(s.dataDir, "project_*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list project files: %w", err)
	}

	var projects []*models.Project
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue // Skip files that can't be read
		}

		var project models.Project
		if err := json.Unmarshal(data, &project); err != nil {
			continue // Skip files that can't be unmarshaled
		}

		projects = append(projects, &project)
	}

	return projects, nil
}

// DeleteProject removes a project from storage
func (s *JSONProjectStore) DeleteProject(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filename := filepath.Join(s.dataDir, fmt.Sprintf("project_%s.json", id))

	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete project file: %w", err)
	}

	return nil
}
