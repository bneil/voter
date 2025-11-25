package project

import (
	"fmt"
	"time"

	"github.com/bneil/voter/internal/models"
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

// AdvanceProject advances the project to the next decision based on the current winner
func (pm *ProgressionManager) AdvanceProject(projectID string, nextDecisionDesc string, nextOptions []string) (*models.Decision, error) {
	project, err := pm.service.GetProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if !project.CanAcceptVotes() {
		return nil, ErrProjectNotActive
	}

	// Check if current decision is complete
	currentDecision := project.GetCurrentDecision()
	if currentDecision != nil && currentDecision.State == models.DecisionStateVoting {
		return nil, fmt.Errorf("current decision is still active")
	}

	// Check if project should end
	if pm.shouldEndProject(project) {
		return nil, pm.service.EndProject(projectID)
	}

	// Start next decision
	decisionID := fmt.Sprintf("decision_%d", project.CurrentTurn+1)
	decision, err := pm.service.StartDecision(projectID, decisionID, nextDecisionDesc, nextOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to start next decision: %w", err)
	}

	return decision, nil
}

// shouldEndProject determines if the project should end based on various conditions
func (pm *ProgressionManager) shouldEndProject(project *models.Project) bool {
	// End if max turns reached
	if project.CurrentTurn >= project.MaxTurns {
		return true
	}

	// End if project is in a terminal state
	if project.IsComplete() {
		return true
	}

	// Could add more complex logic here based on project-specific rules
	// For example, if a certain score is reached, or if no progress is being made

	return false
}

// CalculateProjectScore calculates the overall score for a completed project
func (pm *ProgressionManager) CalculateProjectScore(project *models.Project) int {
	if !project.IsComplete() {
		return 0
	}

	score := 0

	// Score based on number of completed decisions
	score += project.Metrics.TotalDecisions * 10

	// Bonus for faster consensus (lower average time)
	if project.Metrics.AverageConsensusTime > 0 {
		avgSeconds := project.Metrics.AverageConsensusTime.Seconds()
		if avgSeconds < 30 {
			score += 50
		} else if avgSeconds < 60 {
			score += 25
		}
	}

	// Score based on total votes (indicating participation)
	score += project.Metrics.TotalVotes * 2

	return score
}

// GetProjectProgress returns detailed progress information for a project
func (pm *ProgressionManager) GetProjectProgress(projectID string) (*ProjectProgress, error) {
	project, err := pm.service.GetProject(projectID)
	if err != nil {
		return nil, err
	}

	progress := &ProjectProgress{
		ProjectID:          project.ID,
		CurrentTurn:        project.CurrentTurn,
		MaxTurns:           project.MaxTurns,
		State:              project.State,
		TotalDecisions:     project.Metrics.TotalDecisions,
		CompletedDecisions: 0,
		ActiveDecisions:    0,
		Decisions:          make([]DecisionProgress, 0, len(project.Decisions)),
	}

	for _, decision := range project.Decisions {
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

// ProjectProgress represents detailed progress information for a project
type ProjectProgress struct {
	ProjectID          string              `json:"project_id"`
	CurrentTurn        int                 `json:"current_turn"`
	MaxTurns           int                 `json:"max_turns"`
	State              models.ProjectState `json:"state"`
	TotalDecisions     int                 `json:"total_decisions"`
	CompletedDecisions int                 `json:"completed_decisions"`
	ActiveDecisions    int                 `json:"active_decisions"`
	ProgressPercentage float64             `json:"progress_percentage"`
	Decisions          []DecisionProgress  `json:"decisions"`
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
