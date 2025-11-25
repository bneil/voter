package project

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"voter/internal/models"
	"voter/internal/storage"
)

var (
	ErrProjectNotFound  = errors.New("project not found")
	ErrProjectNotActive = errors.New("project is not active")
	ErrInvalidDecision  = errors.New("invalid decision")
	ErrDecisionNotFound = errors.New("decision not found")
	ErrVotingClosed     = errors.New("voting is closed")
	ErrInvalidOption    = errors.New("invalid voting option")
)

// Service manages project sessions and voting logic
type Service struct {
	store  storage.ProjectStore
	voting *VotingService
	mu     sync.RWMutex
}

// NewService creates a new project service
func NewService(store storage.ProjectStore, voting *VotingService) *Service {
	return &Service{
		store:  store,
		voting: voting,
	}
}

// CreateProject creates a new project session
func (s *Service) CreateProject(id, name string, k, maxTurns int) (*models.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	project := models.NewProject(id, name, k, maxTurns)

	if err := s.store.SaveProject(project); err != nil {
		return nil, fmt.Errorf("failed to save project: %w", err)
	}

	return project, nil
}

// GetProject retrieves a project by ID
func (s *Service) GetProject(id string) (*models.Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.store.GetProject(id)
}

// ListProjects returns all projects
func (s *Service) ListProjects() ([]*models.Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.store.ListProjects()
}

// StartDecision starts a new voting decision for a project
func (s *Service) StartDecision(projectID, decisionID, description string, options []string) (*models.Decision, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	project, err := s.store.GetProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if !project.CanAcceptVotes() {
		return nil, ErrProjectNotActive
	}

	// Check if there's already an active decision
	if project.GetCurrentDecision() != nil {
		return nil, errors.New("project already has an active decision")
	}

	decision := models.NewDecision(decisionID, projectID, description, project.CurrentTurn+1, options)
	project.Decisions = append(project.Decisions, *decision)
	project.CurrentTurn = decision.TurnNumber
	project.UpdatedAt = time.Now()

	if err := s.store.SaveProject(project); err != nil {
		return nil, fmt.Errorf("failed to save project: %w", err)
	}

	return decision, nil
}

// CastVote casts a vote for a decision
func (s *Service) CastVote(projectID, decisionID, agentID, option string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	project, err := s.store.GetProject(projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	if !project.CanAcceptVotes() {
		return ErrProjectNotActive
	}

	decision := project.GetCurrentDecision()
	if decision == nil || decision.ID != decisionID {
		return ErrDecisionNotFound
	}

	if decision.State != models.DecisionStateVoting {
		return ErrVotingClosed
	}

	// Cast the vote
	if err := s.voting.CastVote(decision, agentID, option); err != nil {
		return err
	}

	// Check for winner
	if winner := decision.CheckWinner(project.K); winner != nil {
		decision.State = models.DecisionStateCompleted
		decision.Winner = winner
		now := time.Now()
		decision.CompletedAt = &now

		// Update project metrics
		project.Metrics.TotalDecisions++
		project.Metrics.TotalVotes += s.getTotalVotes(decision)
		if project.Metrics.TotalDecisions > 0 {
			// Calculate average consensus time
			totalTime := time.Duration(0)
			for _, d := range project.Decisions {
				if d.CompletedAt != nil {
					totalTime += d.CompletedAt.Sub(d.VotingStarted)
				}
			}
			project.Metrics.AverageConsensusTime = totalTime / time.Duration(project.Metrics.TotalDecisions)
		}
	}

	project.UpdatedAt = time.Now()

	if err := s.store.SaveProject(project); err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	return nil
}

// EndProject ends a project session
func (s *Service) EndProject(projectID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	project, err := s.store.GetProject(projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	if project.IsComplete() {
		return errors.New("project is already complete")
	}

	project.State = models.ProjectStateCompleted
	now := time.Now()
	project.CompletedAt = &now
	project.UpdatedAt = now

	// Close any active decision
	if decision := project.GetCurrentDecision(); decision != nil {
		decision.State = models.DecisionStateCancelled
	}

	if err := s.store.SaveProject(project); err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	return nil
}

// GetProjectStatus returns the current status of a project
func (s *Service) GetProjectStatus(projectID string) (*ProjectStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	project, err := s.store.GetProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	status := &ProjectStatus{
		Project:  project,
		IsActive: project.CanAcceptVotes(),
	}

	if decision := project.GetCurrentDecision(); decision != nil {
		status.CurrentDecision = decision
		status.VoteCounts = make(map[string]int)
		for option, count := range decision.Votes {
			status.VoteCounts[option] = count
		}
	}

	return status, nil
}

// ProjectStatus represents the current status of a project
type ProjectStatus struct {
	Project         *models.Project  `json:"project"`
	IsActive        bool             `json:"is_active"`
	CurrentDecision *models.Decision `json:"current_decision,omitempty"`
	VoteCounts      map[string]int   `json:"vote_counts,omitempty"`
}

// getTotalVotes returns the total number of votes cast for a decision
func (s *Service) getTotalVotes(decision *models.Decision) int {
	total := 0
	for _, count := range decision.Votes {
		total += count
	}
	return total
}
