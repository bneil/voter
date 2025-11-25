package voting

import (
	"fmt"
	"sync"
	"time"

	"github.com/bneil/voter/internal/models"
)

// EnhancedVotingService provides advanced voting capabilities with strategies
type EnhancedVotingService struct {
	votingService  *VotingService
	strategicVoter *StrategicVoter
	mu             sync.RWMutex
}

// VotingService represents the core voting functionality
type VotingService struct {
	mu sync.RWMutex
}

// NewEnhancedVotingService creates a new enhanced voting service
func NewEnhancedVotingService() *EnhancedVotingService {
	return &EnhancedVotingService{
		votingService:  &VotingService{},
		strategicVoter: NewStrategicVoter(),
	}
}

// InitializeStrategies sets up the available voting strategies
func (evs *EnhancedVotingService) InitializeStrategies() {
	evs.strategicVoter.RegisterStrategy("random", NewRandomStrategy())
	evs.strategicVoter.RegisterStrategy("consensus", NewConsensusStrategy())
	evs.strategicVoter.RegisterStrategy("optimal", NewOptimalStrategy("general"))
}

// CastStrategicVote casts a vote using a specific strategy
func (evs *EnhancedVotingService) CastStrategicVote(project *models.Project, decision *models.Decision, agentID, strategyName string) error {
	evs.mu.Lock()
	defer evs.mu.Unlock()

	if decision.State != models.DecisionStateVoting {
		return fmt.Errorf("voting is closed for this decision")
	}

	// Use strategy to decide vote
	chosenOption := evs.strategicVoter.DecideVote(strategyName, project, decision, agentID)

	// Cast the vote
	if !decision.AddVote(chosenOption) {
		return fmt.Errorf("invalid voting option: %s", chosenOption)
	}

	return nil
}

// SimulateAgentVoting simulates multiple agents voting using different strategies
func (evs *EnhancedVotingService) SimulateAgentVoting(project *models.Project, decision *models.Decision, agentCount int) error {
	evs.mu.Lock()
	defer evs.mu.Unlock()

	strategies := []string{"random", "consensus", "optimal"}

	for i := 0; i < agentCount; i++ {
		agentID := fmt.Sprintf("agent_%d", i)
		strategy := strategies[i%len(strategies)]

		chosenOption := evs.strategicVoter.DecideVote(strategy, project, decision, agentID)

		if !decision.AddVote(chosenOption) {
			return fmt.Errorf("failed to cast vote for agent %s: invalid option %s", agentID, chosenOption)
		}
	}

	return nil
}

// AnalyzeVotingPatterns analyzes voting patterns in a completed decision
func (evs *EnhancedVotingService) AnalyzeVotingPatterns(decision *models.Decision) *VotingAnalysis {
	evs.mu.RLock()
	defer evs.mu.RUnlock()

	analysis := &VotingAnalysis{
		TotalVotes:       0,
		OptionVotes:      make(map[string]int),
		VoteDistribution: make(map[string]float64),
	}

	// Count total votes and per-option votes
	for option, count := range decision.Votes {
		analysis.OptionVotes[option] = count
		analysis.TotalVotes += count
	}

	// Calculate vote distribution
	if analysis.TotalVotes > 0 {
		for option, count := range decision.Votes {
			analysis.VoteDistribution[option] = float64(count) / float64(analysis.TotalVotes)
		}
	}

	// Determine consensus strength
	if decision.Winner != nil {
		winnerVotes := decision.Votes[*decision.Winner]
		analysis.ConsensusStrength = float64(winnerVotes) / float64(analysis.TotalVotes)

		// Calculate how many votes ahead the winner was
		maxOtherVotes := 0
		for option, votes := range decision.Votes {
			if option != *decision.Winner && votes > maxOtherVotes {
				maxOtherVotes = votes
			}
		}
		analysis.VotesAhead = winnerVotes - maxOtherVotes
	}

	return analysis
}

// GetVotingRecommendations provides recommendations for improving voting outcomes
func (evs *EnhancedVotingService) GetVotingRecommendations(project *models.Project, decision *models.Decision) []string {
	var recommendations []string

	// Analyze current voting state
	analysis := evs.AnalyzeVotingPatterns(decision)

	// Check if voting is taking too long
	if decision.CompletedAt == nil {
		votingDuration := time.Since(decision.VotingStarted)
		if votingDuration > 5*time.Minute {
			recommendations = append(recommendations, "Consider reducing the K-ahead threshold to speed up consensus")
		}
	}

	// Check vote distribution
	if analysis.TotalVotes > 0 {
		// If votes are too evenly distributed, suggest strategies
		maxPercentage := 0.0
		for _, percentage := range analysis.VoteDistribution {
			if percentage > maxPercentage {
				maxPercentage = percentage
			}
		}

		if maxPercentage < 0.4 {
			recommendations = append(recommendations, "Votes are evenly distributed - consider using consensus-building strategies")
		}
	}

	// Project-specific recommendations
	if project.Name == "Tower of Hanoi" {
		recommendations = append(recommendations, "For Tower of Hanoi, consider optimal strategies that prioritize moving smaller disks")
	}

	return recommendations
}

// VotingAnalysis represents analysis of voting patterns
type VotingAnalysis struct {
	TotalVotes        int                `json:"total_votes"`
	OptionVotes       map[string]int     `json:"option_votes"`
	VoteDistribution  map[string]float64 `json:"vote_distribution"`
	ConsensusStrength float64            `json:"consensus_strength"`
	VotesAhead        int                `json:"votes_ahead"`
}
