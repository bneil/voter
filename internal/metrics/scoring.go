package metrics

import (
	"math"
	"time"

	"voter/internal/models"
)

// Scorer calculates performance scores for games and decisions
type Scorer struct{}

// NewScorer creates a new scorer
func NewScorer() *Scorer {
	return &Scorer{}
}

// CalculateProjectScore calculates the overall score for a completed project
func (s *Scorer) CalculateProjectScore(project *models.Project) *GameScore {
	if !project.IsComplete() {
		return nil
	}

	score := &GameScore{
		GameID:                   project.ID,
		TotalScore:               0,
		CompletionBonus:          0,
		EfficiencyBonus:          0,
		ParticipationBonus:       0,
		QualityScore:             0,
		SpeedScore:               0,
		ConsensusScore:           0,
		GameAverageConsensusTime: project.Metrics.AverageConsensusTime, // Populate new field
	}

	score.TotalScore = project.Metrics.TotalDecisions * 10

	// Completion bonus for finishing the project
	if project.State == models.ProjectStateCompleted {
		score.CompletionBonus = 100
		score.TotalScore += score.CompletionBonus
	}

	if project.Metrics.AverageConsensusTime > 0 {
		avgSeconds := project.Metrics.AverageConsensusTime.Seconds()
		if avgSeconds < 30 {
			score.EfficiencyBonus = 50
		} else if avgSeconds < 60 {
			score.EfficiencyBonus = 25
		}
		score.TotalScore += score.EfficiencyBonus
	}

	score.ParticipationBonus = project.Metrics.TotalVotes * 2
	score.TotalScore += score.ParticipationBonus

	score.QualityScore = s.calculateQualityScore(project)
	score.TotalScore += int(score.QualityScore)

	score.SpeedScore = s.calculateSpeedScore(project)
	score.TotalScore += int(score.SpeedScore)

	score.ConsensusScore = s.calculateConsensusScore(project)
	score.TotalScore += int(score.ConsensusScore)

	return score
}

// CalculateDecisionScore calculates the score for a single decision
func (s *Scorer) CalculateDecisionScore(decision *models.Decision, k int) *DecisionScore {
	score := &DecisionScore{
		DecisionID:        decision.ID,
		ConsensusSpeed:    0,
		ConsensusStrength: 0,
		VoteEfficiency:    0,
		TotalScore:        0,
	}

	if decision.CompletedAt == nil {
		return score
	}

	// Consensus speed (lower time = higher score)
	consensusTime := decision.CompletedAt.Sub(decision.VotingStarted)
	score.ConsensusSpeed = s.calculateTimeScore(consensusTime)

	// Consensus strength (how decisive the winner was)
	if decision.Winner != nil {
		winnerVotes := decision.Votes[*decision.Winner]
		totalVotes := 0
		maxOtherVotes := 0

		for option, votes := range decision.Votes {
			totalVotes += votes
			if option != *decision.Winner && votes > maxOtherVotes {
				maxOtherVotes = votes
			}
		}

		if totalVotes > 0 {
			score.ConsensusStrength = float64(winnerVotes) / float64(totalVotes)

			// Bonus for being ahead by more than K
			votesAhead := winnerVotes - maxOtherVotes
			if votesAhead > k {
				score.ConsensusStrength += 0.1 // Small bonus for decisive wins
			}
		}
	}

	// Vote efficiency (how quickly consensus was reached relative to total possible votes)
	totalPossibleVotes := len(decision.Options) * 10 // Assume max 10 votes per option for efficiency calc
	if totalPossibleVotes > 0 {
		actualVotes := 0
		for _, votes := range decision.Votes {
			actualVotes += votes
		}
		score.VoteEfficiency = math.Min(1.0, float64(actualVotes)/float64(totalPossibleVotes))
	}

	// Calculate total score
	score.TotalScore = (score.ConsensusSpeed * 0.4) +
		(score.ConsensusStrength * 0.4) +
		(score.VoteEfficiency * 0.2)

	return score
}

// calculateQualityScore calculates overall project quality
func (s *Scorer) calculateQualityScore(project *models.Project) float64 {
	if project.Metrics.TotalDecisions == 0 {
		return 0
	}

	// Quality based on consistency of consensus times
	var times []time.Duration
	for _, decision := range project.Decisions {
		if decision.CompletedAt != nil {
			times = append(times, decision.CompletedAt.Sub(decision.VotingStarted))
		}
	}

	if len(times) < 2 {
		return 0.5 // Neutral score for insufficient data
	}

	// Calculate standard deviation of consensus times
	mean := s.calculateMeanDuration(times)
	variance := 0.0
	for _, t := range times {
		diff := t.Seconds() - mean
		variance += diff * diff
	}
	variance /= float64(len(times))
	stdDev := math.Sqrt(variance)

	// Lower standard deviation = higher quality (more consistent)
	maxExpectedStdDev := 60.0 // 1 minute
	quality := math.Max(0, 1.0-(stdDev/maxExpectedStdDev))

	return math.Min(1.0, quality)
}

// calculateSpeedScore calculates how quickly the project reached consensus
func (s *Scorer) calculateSpeedScore(project *models.Project) float64 {
	if project.Metrics.AverageConsensusTime == 0 {
		return 0
	}

	// Score based on average consensus time
	avgSeconds := project.Metrics.AverageConsensusTime.Seconds()

	// Ideal time is around 30 seconds
	idealTime := 30.0
	deviation := math.Abs(avgSeconds - idealTime)

	// Score decreases as deviation from ideal increases
	score := math.Max(0, 1.0-(deviation/60.0)) // 60 second range

	return math.Min(1.0, score)
}

// calculateConsensusScore calculates the strength of consensus across decisions
func (s *Scorer) calculateConsensusScore(project *models.Project) float64 {
	if len(project.Decisions) == 0 {
		return 0
	}

	totalStrength := 0.0
	completedDecisions := 0

	for _, decision := range project.Decisions {
		if decision.Winner != nil && decision.CompletedAt != nil {
			winnerVotes := decision.Votes[*decision.Winner]
			totalVotes := 0

			for _, votes := range decision.Votes {
				totalVotes += votes
			}

			if totalVotes > 0 {
				strength := float64(winnerVotes) / float64(totalVotes)
				totalStrength += strength
				completedDecisions++
			}
		}
	}

	if completedDecisions == 0 {
		return 0
	}

	return totalStrength / float64(completedDecisions)
}

// calculateTimeScore converts consensus time to a score (0-1)
func (s *Scorer) calculateTimeScore(duration time.Duration) float64 {
	seconds := duration.Seconds()

	// Ideal time: 10-30 seconds = score of 1.0
	// Very fast (< 5 seconds) or very slow (> 2 minutes) = lower scores
	if seconds < 5 {
		return 0.7 // Too fast might indicate rushed decisions
	} else if seconds <= 30 {
		return 1.0 // Ideal range
	} else if seconds <= 60 {
		return 0.8 // Acceptable
	} else if seconds <= 120 {
		return 0.5 // Slow
	} else {
		return 0.2 // Very slow
	}
}

// calculateMeanDuration calculates the mean of a slice of durations
func (s *Scorer) calculateMeanDuration(durations []time.Duration) float64 {
	if len(durations) == 0 {
		return 0
	}

	total := 0.0
	for _, d := range durations {
		total += d.Seconds()
	}

	return total / float64(len(durations))
}

// GameScore represents the scoring breakdown for a game
type GameScore struct {
	GameID                   string        `json:"game_id"`
	TotalScore               int           `json:"total_score"`
	CompletionBonus          int           `json:"completion_bonus"`
	EfficiencyBonus          int           `json:"efficiency_bonus"`
	ParticipationBonus       int           `json:"participation_bonus"`
	QualityScore             float64       `json:"quality_score"`
	SpeedScore               float64       `json:"speed_score"`
	ConsensusScore           float64       `json:"consensus_score"`
	GameAverageConsensusTime time.Duration `json:"game_average_consensus_time"` // New field
}

// DecisionScore represents the scoring breakdown for a decision
type DecisionScore struct {
	DecisionID        string  `json:"decision_id"`
	ConsensusSpeed    float64 `json:"consensus_speed"`
	ConsensusStrength float64 `json:"consensus_strength"`
	VoteEfficiency    float64 `json:"vote_efficiency"`
	TotalScore        float64 `json:"total_score"`
}
