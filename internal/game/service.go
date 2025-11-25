package game

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"voter/internal/models"
	"voter/internal/storage"
)

var (
	ErrGameNotFound     = errors.New("game not found")
	ErrGameNotActive    = errors.New("game is not active")
	ErrInvalidDecision  = errors.New("invalid decision")
	ErrDecisionNotFound = errors.New("decision not found")
	ErrVotingClosed     = errors.New("voting is closed")
	ErrInvalidOption    = errors.New("invalid voting option")
)

// Service manages game sessions and voting logic
type Service struct {
	store  storage.GameStore
	voting *VotingService
	mu     sync.RWMutex
}

// NewService creates a new game service
func NewService(store storage.GameStore, voting *VotingService) *Service {
	return &Service{
		store:  store,
		voting: voting,
	}
}

// CreateGame creates a new game session
func (s *Service) CreateGame(id, name string, k, maxTurns int) (*models.Game, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	game := models.NewGame(id, name, k, maxTurns)

	if err := s.store.SaveGame(game); err != nil {
		return nil, fmt.Errorf("failed to save game: %w", err)
	}

	return game, nil
}

// GetGame retrieves a game by ID
func (s *Service) GetGame(id string) (*models.Game, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.store.GetGame(id)
}

// ListGames returns all games
func (s *Service) ListGames() ([]*models.Game, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.store.ListGames()
}

// StartDecision starts a new voting decision for a game
func (s *Service) StartDecision(gameID, decisionID, description string, options []string) (*models.Decision, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	game, err := s.store.GetGame(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	if !game.CanAcceptVotes() {
		return nil, ErrGameNotActive
	}

	// Check if there's already an active decision
	if game.GetCurrentDecision() != nil {
		return nil, errors.New("game already has an active decision")
	}

	decision := models.NewDecision(decisionID, gameID, description, game.CurrentTurn+1, options)
	game.Decisions = append(game.Decisions, *decision)
	game.CurrentTurn = decision.TurnNumber
	game.UpdatedAt = time.Now()

	if err := s.store.SaveGame(game); err != nil {
		return nil, fmt.Errorf("failed to save game: %w", err)
	}

	return decision, nil
}

// CastVote casts a vote for a decision
func (s *Service) CastVote(gameID, decisionID, agentID, option string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	game, err := s.store.GetGame(gameID)
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	if !game.CanAcceptVotes() {
		return ErrGameNotActive
	}

	decision := game.GetCurrentDecision()
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
	if winner := decision.CheckWinner(game.K); winner != nil {
		decision.State = models.DecisionStateCompleted
		decision.Winner = winner
		now := time.Now()
		decision.CompletedAt = &now

		// Update game metrics
		game.Metrics.TotalDecisions++
		game.Metrics.TotalVotes += s.getTotalVotes(decision)
		if game.Metrics.TotalDecisions > 0 {
			// Calculate average consensus time
			totalTime := time.Duration(0)
			for _, d := range game.Decisions {
				if d.CompletedAt != nil {
					totalTime += d.CompletedAt.Sub(d.VotingStarted)
				}
			}
			game.Metrics.AverageConsensusTime = totalTime / time.Duration(game.Metrics.TotalDecisions)
		}
	}

	game.UpdatedAt = time.Now()

	if err := s.store.SaveGame(game); err != nil {
		return fmt.Errorf("failed to save game: %w", err)
	}

	return nil
}

// EndGame ends a game session
func (s *Service) EndGame(gameID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	game, err := s.store.GetGame(gameID)
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	if game.IsComplete() {
		return errors.New("game is already complete")
	}

	game.State = models.GameStateCompleted
	now := time.Now()
	game.CompletedAt = &now
	game.UpdatedAt = now

	// Close any active decision
	if decision := game.GetCurrentDecision(); decision != nil {
		decision.State = models.DecisionStateCancelled
	}

	if err := s.store.SaveGame(game); err != nil {
		return fmt.Errorf("failed to save game: %w", err)
	}

	return nil
}

// GetGameStatus returns the current status of a game
func (s *Service) GetGameStatus(gameID string) (*GameStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	game, err := s.store.GetGame(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	status := &GameStatus{
		Game:     game,
		IsActive: game.CanAcceptVotes(),
	}

	if decision := game.GetCurrentDecision(); decision != nil {
		status.CurrentDecision = decision
		status.VoteCounts = make(map[string]int)
		for option, count := range decision.Votes {
			status.VoteCounts[option] = count
		}
	}

	return status, nil
}

// GameStatus represents the current status of a game
type GameStatus struct {
	Game            *models.Game     `json:"game"`
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
