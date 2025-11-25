package models

import (
	"time"
)

// ProjectState represents the current state of a project session
type ProjectState string

const (
	ProjectStateActive    ProjectState = "active"
	ProjectStatePaused    ProjectState = "paused"
	ProjectStateCompleted ProjectState = "completed"
	ProjectStateCancelled ProjectState = "cancelled"
)

// Project represents a complete project session
type Project struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	State       ProjectState   `json:"state"`
	K           int            `json:"k"`         // K-ahead threshold
	MaxTurns    int            `json:"max_turns"` // Maximum number of turns
	CurrentTurn int            `json:"current_turn"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
	Score       int            `json:"score"` // Overall project score
	Metrics     ProjectMetrics `json:"metrics"`
	Decisions   []Decision     `json:"decisions"`
}

// Decision represents a single voting decision within a project
type Decision struct {
	ID            string         `json:"id"`
	ProjectID     string         `json:"project_id"`
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

// ProjectMetrics tracks performance metrics for the project
type ProjectMetrics struct {
	TotalDecisions       int           `json:"total_decisions"`
	AverageConsensusTime time.Duration `json:"average_consensus_time"`
	TotalVotes           int           `json:"total_votes"`
}

// Vote represents a single vote cast by an agent
type Vote struct {
	ID         string    `json:"id"`
	DecisionID string    `json:"decision_id"`
	ProjectID  string    `json:"project_id"`
	AgentID    string    `json:"agent_id"`
	Option     string    `json:"option"`
	Timestamp  time.Time `json:"timestamp"`
}

// NewProject creates a new project with the given parameters
func NewProject(id, name string, k, maxTurns int) *Project {
	now := time.Now()
	return &Project{
		ID:          id,
		Name:        name,
		State:       ProjectStateActive,
		K:           k,
		MaxTurns:    maxTurns,
		CurrentTurn: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
		Score:       0,
		Metrics: ProjectMetrics{
			TotalDecisions:       0,
			AverageConsensusTime: 0,
			TotalVotes:           0,
		},
		Decisions: []Decision{},
	}
}

// NewDecision creates a new decision for a project
func NewDecision(id, projectID, description string, turnNumber int, options []string) *Decision {
	now := time.Now()
	votes := make(map[string]int)
	for _, option := range options {
		votes[option] = 0
	}

	return &Decision{
		ID:            id,
		ProjectID:     projectID,
		TurnNumber:    turnNumber,
		Description:   description,
		Options:       options,
		State:         DecisionStateVoting,
		Votes:         votes,
		CreatedAt:     now,
		VotingStarted: now,
	}
}

// IsComplete checks if the project is in a terminal state
func (p *Project) IsComplete() bool {
	return p.State == ProjectStateCompleted || p.State == ProjectStateCancelled
}

// CanAcceptVotes checks if the project can accept new votes
func (p *Project) CanAcceptVotes() bool {
	return p.State == ProjectStateActive
}

// GetCurrentDecision returns the current active decision, if any
func (p *Project) GetCurrentDecision() *Decision {
	for i := len(p.Decisions) - 1; i >= 0; i-- {
		if p.Decisions[i].State == DecisionStateVoting {
			return &p.Decisions[i]
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
