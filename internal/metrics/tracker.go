package metrics

import (
	"sync"
	"time"

	"github.com/bneil/voter/internal/models"
)

// Tracker tracks and analyzes project metrics over time
type Tracker struct {
	mu             sync.RWMutex
	projectScores  map[string]*GameScore
	decisionScores map[string]*DecisionScore
	globalStats    *GlobalStats
}

// GlobalStats represents global statistics across all projects
type GlobalStats struct {
	TotalProjects        int                       `json:"total_projects"`
	TotalDecisions       int                       `json:"total_decisions"`
	AverageProjectScore  float64                   `json:"average_project_score"`
	AverageConsensusTime time.Duration             `json:"average_consensus_time"`
	BestProjectScore     int                       `json:"best_project_score"`
	BestProjectID        string                    `json:"best_project_id"`
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
		projectScores:  make(map[string]*GameScore),
		decisionScores: make(map[string]*DecisionScore),
		globalStats: &GlobalStats{
			StrategyPerformance: make(map[string]*StrategyStats),
		},
	}
}

// RecordProjectScore records the score for a completed project
func (t *Tracker) RecordProjectScore(project *models.Project, score *GameScore) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.projectScores[project.ID] = score
	t.updateGlobalStats(project, score)
}

// RecordDecisionScore records the score for a completed decision
func (t *Tracker) RecordDecisionScore(decisionID string, score *DecisionScore) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.decisionScores[decisionID] = score
}

// GetProjectScore retrieves the score for a specific project
func (t *Tracker) GetProjectScore(projectID string) *GameScore {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.projectScores[projectID]
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

// GetTopProjects returns the top N projects by score
func (t *Tracker) GetTopProjects(limit int) []*GameScore {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var scores []*GameScore
	for _, score := range t.projectScores {
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
		ProjectScores:  make([]int, 0, len(t.projectScores)),
		ConsensusTimes: make([]time.Duration, 0, len(t.projectScores)),
	}

	for _, score := range t.projectScores {
		trends.ProjectScores = append(trends.ProjectScores, score.TotalScore)
	}

	// Calculate trends
	if len(trends.ProjectScores) > 1 {
		trends.ScoreTrend = t.calculateTrend(trends.ProjectScores)
	}

	return trends
}

// updateGlobalStats updates the global statistics with new game data
func (t *Tracker) updateGlobalStats(project *models.Project, score *GameScore) {
	t.globalStats.TotalProjects++
	t.globalStats.TotalDecisions += project.Metrics.TotalDecisions

	// Update average project score
	totalScoreSum := 0
	for _, s := range t.projectScores {
		totalScoreSum += s.TotalScore
	}
	if t.globalStats.TotalProjects > 0 {
		t.globalStats.AverageProjectScore = float64(totalScoreSum) / float64(t.globalStats.TotalProjects)
	} else {
		t.globalStats.AverageProjectScore = 0
	}

	// Update best project
	if score.TotalScore > t.globalStats.BestProjectScore {
		t.globalStats.BestProjectScore = score.TotalScore
		t.globalStats.BestProjectID = project.ID
	}

	// Update average consensus time
	totalConsensusTimeNanos := int64(0)
	projectsWithConsensusTime := 0
	for _, s := range t.projectScores {
		if s.GameAverageConsensusTime > 0 {
			totalConsensusTimeNanos += s.GameAverageConsensusTime.Nanoseconds()
			projectsWithConsensusTime++
		}
	}
	if projectsWithConsensusTime > 0 {
		t.globalStats.AverageConsensusTime = time.Duration(totalConsensusTimeNanos / int64(projectsWithConsensusTime))
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
	ProjectScores  []int           `json:"project_scores"`
	ConsensusTimes []time.Duration `json:"consensus_times"`
	ScoreTrend     string          `json:"score_trend"`
}
