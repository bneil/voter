package project

import (
	"errors"
	"sync"
	"time"

	"voter/internal/models"
)

var (
	ErrInvalidVote = errors.New("invalid vote")
)

// VotingService handles voting operations with atomicity
type VotingService struct {
	mu sync.RWMutex
}

// NewVotingService creates a new voting service
func NewVotingService() *VotingService {
	return &VotingService{}
}

// CastVote casts a vote for a decision with atomic operations
func (vs *VotingService) CastVote(decision *models.Decision, agentID, option string) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	if decision.State != models.DecisionStateVoting {
		return ErrVotingClosed
	}

	if !decision.AddVote(option) {
		return ErrInvalidVote
	}

	return nil
}

// GetVoteCounts returns the current vote counts for a decision
func (vs *VotingService) GetVoteCounts(decision *models.Decision) map[string]int {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	counts := make(map[string]int)
	for option, count := range decision.Votes {
		counts[option] = count
	}
	return counts
}

// CreateVote creates a vote record
func (vs *VotingService) CreateVote(id, decisionID, projectID, agentID, option string) *models.Vote {
	return &models.Vote{
		ID:         id,
		DecisionID: decisionID,
		ProjectID:  projectID,
		AgentID:    agentID,
		Option:     option,
		Timestamp:  time.Now(),
	}
}
