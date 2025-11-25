package models

import (
	"time"
)

// GameState represents the current state of a game session
type GameState string

const (
	GameStateActive    GameState = "active"
	GameStatePaused    GameState = "paused"
	GameStateCompleted GameState = "completed"
	GameStateCancelled GameState = "cancelled"
)

// Game represents a complete game session
type Game struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	State       GameState   `json:"state"`
	K           int         `json:"k"`         // K-ahead threshold
	MaxTurns    int         `json:"max_turns"` // Maximum number of turns
	CurrentTurn int         `json:"current_turn"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	CompletedAt *time.Time  `json:"completed_at,omitempty"`
	Score       int         `json:"score"` // Overall game score
	Metrics     GameMetrics `json:"metrics"`
	Decisions   []Decision  `json:"decisions"`
}

// Decision represents a single voting decision within a game
type Decision struct {
	ID            string         `json:"id"`
	GameID        string         `json:"game_id"`
	TurnNumber    int            `json:"turn_number"`
	Description   string         `json:"description"`
	Options       []string       `json:"options"`
	State         DecisionState  `json:"state"`
	Winner        *string        `json:"winner,omitempty"`
	Votes         map[string]int `json:"votes"` // option -> vote count
	CreatedAt     time.Time      `json:"created_at"`
	CompletedAt   *time.Time     `json:"completed_at,omitempty"`
	VotingStarted time.Time      `json:"voting_started"`
}

// DecisionState represents the state of a decision
type DecisionState string

const (
	DecisionStateVoting    DecisionState = "voting"
	DecisionStateCompleted DecisionState = "completed"
	DecisionStateCancelled DecisionState = "cancelled"
)

// GameMetrics tracks performance metrics for the game
type GameMetrics struct {
	TotalDecisions       int           `json:"total_decisions"`
	AverageConsensusTime time.Duration `json:"average_consensus_time"`
	TotalVotes           int           `json:"total_votes"`
}

// Vote represents a single vote cast by an agent
type Vote struct {
	ID         string    `json:"id"`
	DecisionID string    `json:"decision_id"`
	GameID     string    `json:"game_id"`
	AgentID    string    `json:"agent_id"`
	Option     string    `json:"option"`
	Timestamp  time.Time `json:"timestamp"`
}

// NewGame creates a new game with the given parameters
func NewGame(id, name string, k, maxTurns int) *Game {
	now := time.Now()
	return &Game{
		ID:          id,
		Name:        name,
		State:       GameStateActive,
		K:           k,
		MaxTurns:    maxTurns,
		CurrentTurn: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
		Score:       0,
		Metrics: GameMetrics{
			TotalDecisions:       0,
			AverageConsensusTime: 0,
			TotalVotes:           0,
		},
		Decisions: []Decision{},
	}
}

// NewDecision creates a new decision for a game
func NewDecision(id, gameID, description string, turnNumber int, options []string) *Decision {
	now := time.Now()
	votes := make(map[string]int)
	for _, option := range options {
		votes[option] = 0
	}

	return &Decision{
		ID:            id,
		GameID:        gameID,
		TurnNumber:    turnNumber,
		Description:   description,
		Options:       options,
		State:         DecisionStateVoting,
		Votes:         votes,
		CreatedAt:     now,
		VotingStarted: now,
	}
}

// IsComplete checks if the game is in a terminal state
func (g *Game) IsComplete() bool {
	return g.State == GameStateCompleted || g.State == GameStateCancelled
}

// CanAcceptVotes checks if the game can accept new votes
func (g *Game) CanAcceptVotes() bool {
	return g.State == GameStateActive
}

// GetCurrentDecision returns the current active decision, if any
func (g *Game) GetCurrentDecision() *Decision {
	for i := len(g.Decisions) - 1; i >= 0; i-- {
		if g.Decisions[i].State == DecisionStateVoting {
			return &g.Decisions[i]
		}
	}
	return nil
}

// CheckWinner determines if any option has reached the K-ahead threshold
func (d *Decision) CheckWinner(k int) *string {
	if len(d.Votes) == 0 {
		return nil
	}

	// Find the option with the most votes
	var maxOption string
	maxVotes := 0
	for option, votes := range d.Votes {
		if votes > maxVotes {
			maxVotes = votes
			maxOption = option
		}
	}

	// Check if it has K more votes than any other option
	for option, votes := range d.Votes {
		if option != maxOption && maxVotes-votes < k {
			return nil // Not enough of a lead
		}
	}

	return &maxOption
}

// AddVote adds a vote to the decision
func (d *Decision) AddVote(option string) bool {
	if d.State != DecisionStateVoting {
		return false
	}

	if _, exists := d.Votes[option]; !exists {
		return false // Invalid option
	}

	d.Votes[option]++
	return true
}
