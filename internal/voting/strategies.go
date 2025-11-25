package voting

import (
	"math/rand"
	"time"

	"voter/internal/models"
)

// Strategy represents different voting strategies that agents can use
type Strategy interface {
	// DecideVote returns the option an agent should vote for given the current project state
	DecideVote(project *models.Project, decision *models.Decision, agentID string) string
}

// RandomStrategy votes randomly among available options
type RandomStrategy struct {
	rng *rand.Rand
}

func NewRandomStrategy() *RandomStrategy {
	return &RandomStrategy{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *RandomStrategy) DecideVote(project *models.Project, decision *models.Decision, agentID string) string {
	if len(decision.Options) == 0 {
		return ""
	}
	return decision.Options[s.rng.Intn(len(decision.Options))]
}

// ConsensusStrategy tries to vote for options that are gaining consensus
type ConsensusStrategy struct{}

func NewConsensusStrategy() *ConsensusStrategy {
	return &ConsensusStrategy{}
}

func (s *ConsensusStrategy) DecideVote(project *models.Project, decision *models.Decision, agentID string) string {
	if len(decision.Options) == 0 {
		return ""
	}

	// Find the option with the most votes
	maxVotes := 0
	var leadingOption string

	for option, votes := range decision.Votes {
		if votes > maxVotes {
			maxVotes = votes
			leadingOption = option
		}
	}

	// If there's a clear leader, vote for it
	if leadingOption != "" && maxVotes > 0 {
		return leadingOption
	}

	// Otherwise vote randomly
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return decision.Options[rng.Intn(len(decision.Options))]
}

// OptimalStrategy uses game-specific knowledge to make optimal decisions
// This is a placeholder for more sophisticated strategies
type OptimalStrategy struct {
	gameType string
}

func NewOptimalStrategy(gameType string) *OptimalStrategy {
	return &OptimalStrategy{
		gameType: gameType,
	}
}

func (s *OptimalStrategy) DecideVote(project *models.Project, decision *models.Decision, agentID string) string {
	// For Tower of Hanoi, this could implement optimal solving strategies
	switch s.gameType {
	case "tower-of-hanoi":
		return s.decideTowerOfHanoi(project, decision)
	default:
		// Fall back to consensus strategy
		strategy := NewConsensusStrategy()
		return strategy.DecideVote(project, decision, agentID)
	}
}

func (s *OptimalStrategy) decideTowerOfHanoi(project *models.Project, decision *models.Decision) string {
	// This is a simplified Tower of Hanoi strategy
	// In a real implementation, this would analyze the current board state
	// and determine the optimal move based on the game's rules

	// For now, use a simple heuristic: prefer moves that seem more strategic
	if len(decision.Options) >= 3 {
		// In Tower of Hanoi, moving the smallest disk is often optimal
		// This is a placeholder - real implementation would need board state
		return decision.Options[0] // Assume first option is "move smallest disk"
	}

	// Fall back to random
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return decision.Options[rng.Intn(len(decision.Options))]
}

// StrategicVoter manages different voting strategies
type StrategicVoter struct {
	strategies map[string]Strategy
}

func NewStrategicVoter() *StrategicVoter {
	return &StrategicVoter{
		strategies: make(map[string]Strategy),
	}
}

// RegisterStrategy registers a voting strategy
func (sv *StrategicVoter) RegisterStrategy(name string, strategy Strategy) {
	sv.strategies[name] = strategy
}

// GetStrategy returns a strategy by name
func (sv *StrategicVoter) GetStrategy(name string) Strategy {
	if strategy, exists := sv.strategies[name]; exists {
		return strategy
	}
	// Default to random strategy
	return NewRandomStrategy()
}

// DecideVote uses the specified strategy to make a voting decision
func (sv *StrategicVoter) DecideVote(strategyName string, project *models.Project, decision *models.Decision, agentID string) string {
	strategy := sv.GetStrategy(strategyName)
	return strategy.DecideVote(project, decision, agentID)
}

// AdaptiveStrategy adjusts its behavior based on game performance
type AdaptiveStrategy struct {
	strategyHistory map[string][]bool // strategy -> success history
	currentStrategy string
}

func NewAdaptiveStrategy() *AdaptiveStrategy {
	return &AdaptiveStrategy{
		strategyHistory: make(map[string][]bool),
		currentStrategy: "random",
	}
}

func (s *AdaptiveStrategy) DecideVote(project *models.Project, decision *models.Decision, agentID string) string {
	// Choose strategy based on historical performance
	bestStrategy := s.getBestStrategy()

	var strategy Strategy
	switch bestStrategy {
	case "consensus":
		strategy = NewConsensusStrategy()
	case "optimal":
		strategy = NewOptimalStrategy("general")
	default:
		strategy = NewRandomStrategy()
	}

	return strategy.DecideVote(project, decision, agentID)
}

// RecordOutcome records the outcome of a strategy decision
func (s *AdaptiveStrategy) RecordOutcome(strategy string, success bool) {
	if s.strategyHistory[strategy] == nil {
		s.strategyHistory[strategy] = make([]bool, 0)
	}
	s.strategyHistory[strategy] = append(s.strategyHistory[strategy], success)
}

func (s *AdaptiveStrategy) getBestStrategy() string {
	bestStrategy := "random"
	bestScore := 0.0

	for strategy, outcomes := range s.strategyHistory {
		if len(outcomes) == 0 {
			continue
		}

		score := s.calculateScore(outcomes)
		if score > bestScore {
			bestScore = score
			bestStrategy = strategy
		}
	}

	return bestStrategy
}

func (s *AdaptiveStrategy) calculateScore(outcomes []bool) float64 {
	if len(outcomes) == 0 {
		return 0
	}

	successes := 0
	for _, outcome := range outcomes {
		if outcome {
			successes++
		}
	}

	// Weight recent outcomes more heavily
	score := float64(successes) / float64(len(outcomes))

	// Add recency bonus
	if len(outcomes) > 0 && outcomes[len(outcomes)-1] {
		score += 0.1
	}

	return score
}
