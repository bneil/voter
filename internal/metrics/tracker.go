package metrics

import (
	"sync"
	"time"

	"voter/internal/models"
)

// Tracker tracks and analyzes game metrics over time
type Tracker struct {
	mu             sync.RWMutex
	gameScores     map[string]*GameScore
	decisionScores map[string]*DecisionScore
	globalStats    *GlobalStats
}

// GlobalStats represents global statistics across all games
type GlobalStats struct {
	TotalGames           int                       `json:"total_games"`
	TotalDecisions       int                       `json:"total_decisions"`
	AverageGameScore     float64                   `json:"average_game_score"`
	AverageConsensusTime time.Duration             `json:"average_consensus_time"`
	BestGameScore        int                       `json:"best_game_score"`
	BestGameID           string                    `json:"best_game_id"`
	StrategyPerformance  map[string]*StrategyStats `json:"strategy_performance"`
}

// StrategyStats tracks performance of different voting strategies
type StrategyStats struct {
	TotalUses    int           `json:"total_uses"`
	SuccessRate  float64       `json:"success_rate"`
	AverageScore float64       `json:"average_score"`
	AverageTime  time.Duration `json:"average_time"`
}

// NewTracker creates a new metrics tracker
func NewTracker() *Tracker {
	return &Tracker{
		gameScores:     make(map[string]*GameScore),
		decisionScores: make(map[string]*DecisionScore),
		globalStats: &GlobalStats{
			StrategyPerformance: make(map[string]*StrategyStats),
		},
	}
}

// RecordGameScore records the score for a completed game
func (t *Tracker) RecordGameScore(game *models.Game, score *GameScore) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.gameScores[game.ID] = score
	t.updateGlobalStats(game, score)
}

// RecordDecisionScore records the score for a completed decision
func (t *Tracker) RecordDecisionScore(decisionID string, score *DecisionScore) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.decisionScores[decisionID] = score
}

// GetGameScore retrieves the score for a specific game
func (t *Tracker) GetGameScore(gameID string) *GameScore {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.gameScores[gameID]
}

// GetDecisionScore retrieves the score for a specific decision
func (t *Tracker) GetDecisionScore(decisionID string) *DecisionScore {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.decisionScores[decisionID]
}

// GetGlobalStats returns the current global statistics
func (t *Tracker) GetGlobalStats() *GlobalStats {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Return a copy to avoid race conditions
	stats := *t.globalStats
	return &stats
}

// GetTopGames returns the top N games by score
func (t *Tracker) GetTopGames(limit int) []*GameScore {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var scores []*GameScore
	for _, score := range t.gameScores {
		scores = append(scores, score)
	}

	// Sort by total score (descending)
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].TotalScore > scores[i].TotalScore {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	if limit > 0 && len(scores) > limit {
		return scores[:limit]
	}

	return scores
}

// AnalyzeStrategyPerformance analyzes how different strategies perform
func (t *Tracker) AnalyzeStrategyPerformance() map[string]*StrategyStats {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// This would require tracking which strategies were used for each decision
	// For now, return the current strategy performance
	return t.globalStats.StrategyPerformance
}

// GetPerformanceTrends analyzes performance trends over time
func (t *Tracker) GetPerformanceTrends() *PerformanceTrends {
	t.mu.RLock()
	defer t.mu.RUnlock()

	trends := &PerformanceTrends{
		TimeRange:      "all",
		GameScores:     make([]int, 0, len(t.gameScores)),
		ConsensusTimes: make([]time.Duration, 0, len(t.gameScores)),
	}

	for _, score := range t.gameScores {
		trends.GameScores = append(trends.GameScores, score.TotalScore)
	}

	// Calculate trends
	if len(trends.GameScores) > 1 {
		trends.ScoreTrend = t.calculateTrend(trends.GameScores)
	}

	return trends
}

// updateGlobalStats updates the global statistics with new game data
func (t *Tracker) updateGlobalStats(game *models.Game, score *GameScore) {
	t.globalStats.TotalGames++
	t.globalStats.TotalDecisions += game.Metrics.TotalDecisions

	// Update average game score
	totalScoreSum := 0
	for _, s := range t.gameScores {
		totalScoreSum += s.TotalScore
	}
	if t.globalStats.TotalGames > 0 {
		t.globalStats.AverageGameScore = float64(totalScoreSum) / float64(t.globalStats.TotalGames)
	} else {
		t.globalStats.AverageGameScore = 0
	}

	// Update best game
	if score.TotalScore > t.globalStats.BestGameScore {
		t.globalStats.BestGameScore = score.TotalScore
		t.globalStats.BestGameID = game.ID
	}

	// Update average consensus time
	totalConsensusTimeNanos := int64(0)
	gamesWithConsensusTime := 0
	for _, s := range t.gameScores {
		if s.GameAverageConsensusTime > 0 {
			totalConsensusTimeNanos += s.GameAverageConsensusTime.Nanoseconds()
			gamesWithConsensusTime++
		}
	}
	if gamesWithConsensusTime > 0 {
		t.globalStats.AverageConsensusTime = time.Duration(totalConsensusTimeNanos / int64(gamesWithConsensusTime))
	} else {
		t.globalStats.AverageConsensusTime = 0
	}


}

// calculateTrend calculates the trend direction of a series of values
func (t *Tracker) calculateTrend(values []int) string {
	if len(values) < 2 {
		return "insufficient-data"
	}

	// Simple linear trend calculation
	n := len(values)
	sumX := 0
	sumY := 0
	sumXY := 0
	sumXX := 0

	for i, y := range values {
		x := i
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}

	// Calculate slope
	slope := float64(n*sumXY-sumX*sumY) / float64(n*sumXX-sumX*sumX)

	if slope > 1 {
		return "improving"
	} else if slope < -1 {
		return "declining"
	} else {
		return "stable"
	}
}

// PerformanceTrends represents performance trends over time
type PerformanceTrends struct {
	TimeRange      string          `json:"time_range"`
	GameScores     []int           `json:"game_scores"`
	ConsensusTimes []time.Duration `json:"consensus_times"`
	ScoreTrend     string          `json:"score_trend"`
}
