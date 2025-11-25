package game

import (
	"fmt"
	"time"

	"voter/internal/models"
)

// ProgressionManager handles game progression logic
type ProgressionManager struct {
	service *Service
}

// NewProgressionManager creates a new progression manager
func NewProgressionManager(service *Service) *ProgressionManager {
	return &ProgressionManager{
		service: service,
	}
}

// AdvanceGame advances the game to the next decision based on the current winner
func (pm *ProgressionManager) AdvanceGame(gameID string, nextDecisionDesc string, nextOptions []string) (*models.Decision, error) {
	game, err := pm.service.GetGame(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	if !game.CanAcceptVotes() {
		return nil, ErrGameNotActive
	}

	// Check if current decision is complete
	currentDecision := game.GetCurrentDecision()
	if currentDecision != nil && currentDecision.State == models.DecisionStateVoting {
		return nil, fmt.Errorf("current decision is still active")
	}

	// Check if game should end
	if pm.shouldEndGame(game) {
		return nil, pm.service.EndGame(gameID)
	}

	// Start next decision
	decisionID := fmt.Sprintf("decision_%d", game.CurrentTurn+1)
	decision, err := pm.service.StartDecision(gameID, decisionID, nextDecisionDesc, nextOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to start next decision: %w", err)
	}

	return decision, nil
}

// shouldEndGame determines if the game should end based on various conditions
func (pm *ProgressionManager) shouldEndGame(game *models.Game) bool {
	// End if max turns reached
	if game.CurrentTurn >= game.MaxTurns {
		return true
	}

	// End if game is in a terminal state
	if game.IsComplete() {
		return true
	}

	// Could add more complex logic here based on game-specific rules
	// For example, if a certain score is reached, or if no progress is being made

	return false
}

// CalculateGameScore calculates the overall score for a completed game
func (pm *ProgressionManager) CalculateGameScore(game *models.Game) int {
	if !game.IsComplete() {
		return 0
	}

	score := 0

	// Score based on number of completed decisions
	score += game.Metrics.TotalDecisions * 10

	// Bonus for faster consensus (lower average time)
	if game.Metrics.AverageConsensusTime > 0 {
		avgSeconds := game.Metrics.AverageConsensusTime.Seconds()
		if avgSeconds < 30 {
			score += 50
		} else if avgSeconds < 60 {
			score += 25
		}
	}

	// Score based on total votes (indicating participation)
	score += game.Metrics.TotalVotes * 2

	return score
}

// GetGameProgress returns detailed progress information for a game
func (pm *ProgressionManager) GetGameProgress(gameID string) (*GameProgress, error) {
	game, err := pm.service.GetGame(gameID)
	if err != nil {
		return nil, err
	}

	progress := &GameProgress{
		GameID:             game.ID,
		CurrentTurn:        game.CurrentTurn,
		MaxTurns:           game.MaxTurns,
		State:              game.State,
		TotalDecisions:     game.Metrics.TotalDecisions,
		CompletedDecisions: 0,
		ActiveDecisions:    0,
		Decisions:          make([]DecisionProgress, 0, len(game.Decisions)),
	}

	for _, decision := range game.Decisions {
		decisionProgress := DecisionProgress{
			ID:          decision.ID,
			TurnNumber:  decision.TurnNumber,
			Description: decision.Description,
			State:       decision.State,
			Options:     decision.Options,
			VoteCounts:  make(map[string]int),
		}

		for option, count := range decision.Votes {
			decisionProgress.VoteCounts[option] = count
		}

		if decision.Winner != nil {
			decisionProgress.Winner = *decision.Winner
		}

		if decision.CompletedAt != nil {
			decisionProgress.CompletedAt = *decision.CompletedAt
			decisionProgress.ConsensusTime = decision.CompletedAt.Sub(decision.VotingStarted)
		}

		progress.Decisions = append(progress.Decisions, decisionProgress)

		switch decision.State {
		case models.DecisionStateCompleted:
			progress.CompletedDecisions++
		case models.DecisionStateVoting:
			progress.ActiveDecisions++
		}
	}

	progress.ProgressPercentage = float64(progress.CurrentTurn) / float64(progress.MaxTurns) * 100
	if progress.ProgressPercentage > 100 {
		progress.ProgressPercentage = 100
	}

	return progress, nil
}

// GameProgress represents detailed progress information for a game
type GameProgress struct {
	GameID             string             `json:"game_id"`
	CurrentTurn        int                `json:"current_turn"`
	MaxTurns           int                `json:"max_turns"`
	State              models.GameState   `json:"state"`
	TotalDecisions     int                `json:"total_decisions"`
	CompletedDecisions int                `json:"completed_decisions"`
	ActiveDecisions    int                `json:"active_decisions"`
	ProgressPercentage float64            `json:"progress_percentage"`
	Decisions          []DecisionProgress `json:"decisions"`
}

// DecisionProgress represents progress information for a single decision
type DecisionProgress struct {
	ID            string               `json:"id"`
	TurnNumber    int                  `json:"turn_number"`
	Description   string               `json:"description"`
	State         models.DecisionState `json:"state"`
	Options       []string             `json:"options"`
	VoteCounts    map[string]int       `json:"vote_counts"`
	Winner        string               `json:"winner,omitempty"`
	CompletedAt   time.Time            `json:"completed_at,omitempty"`
	ConsensusTime time.Duration        `json:"consensus_time,omitempty"`
}
