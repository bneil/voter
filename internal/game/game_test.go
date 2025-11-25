package game_test

import (
	"testing"

	"voter/internal/game"
	"voter/internal/models"
	"voter/internal/storage"
	"voter/internal/voting"
)

func setupTestServices(t *testing.T) (*game.Service, *storage.JSONGameStore) {
	t.Helper()

	store, err := storage.NewJSONGameStore(t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}

	votingService := game.NewVotingService()
	service := game.NewService(store, votingService)

	return service, store
}

func TestCreateGame(t *testing.T) {
	service, _ := setupTestServices(t)

	game, err := service.CreateGame("test-game", "Test Game", 3, 10)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	if game.ID != "test-game" {
		t.Errorf("Expected ID 'test-game', got '%s'", game.ID)
	}

	if game.K != 3 {
		t.Errorf("Expected K=3, got %d", game.K)
	}

	if game.MaxTurns != 10 {
		t.Errorf("Expected MaxTurns=10, got %d", game.MaxTurns)
	}
}

func TestStartDecision(t *testing.T) {
	service, _ := setupTestServices(t)

	// Create game
	_, err := service.CreateGame("test-game", "Test Game", 3, 10)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Start decision
	options := []string{"option1", "option2", "option3"}
	decision, err := service.StartDecision("test-game", "decision-1", "Test decision", options)
	if err != nil {
		t.Fatalf("Failed to start decision: %v", err)
	}

	if decision.ID != "decision-1" {
		t.Errorf("Expected decision ID 'decision-1', got '%s'", decision.ID)
	}

	if len(decision.Options) != 3 {
		t.Errorf("Expected 3 options, got %d", len(decision.Options))
	}

	if decision.State != models.DecisionStateVoting {
		t.Errorf("Expected decision state 'voting', got '%s'", decision.State)
	}
}

func TestCastVote(t *testing.T) {
	service, _ := setupTestServices(t)

	// Create game with K=2
	_, err := service.CreateGame("test-game", "Test Game", 2, 10)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	options := []string{"A", "B", "C"}
	decision, err := service.StartDecision("test-game", "decision-1", "Test decision", options)
	if err != nil {
		t.Fatalf("Failed to start decision: %v", err)
	}

	// Cast votes
	err = service.CastVote("test-game", decision.ID, "agent1", "A")
	if err != nil {
		t.Fatalf("Failed to cast vote: %v", err)
	}

	err = service.CastVote("test-game", decision.ID, "agent2", "A")
	if err != nil {
		t.Fatalf("Failed to cast vote: %v", err)
	}

	// Check that decision completed with A as winner
	status, err := service.GetGameStatus("test-game")
	if err != nil {
		t.Fatalf("Failed to get game status: %v", err)
	}

	if len(status.Game.Decisions) == 0 {
		t.Fatal("Expected at least one decision in game")
	}

	completedDecision := status.Game.Decisions[len(status.Game.Decisions)-1]
	if completedDecision.Winner == nil || *completedDecision.Winner != "A" {
		t.Errorf("Expected A to be winner, got %v", completedDecision.Winner)
	}

	if completedDecision.State != models.DecisionStateCompleted {
		t.Errorf("Expected decision to be completed, got state %s", completedDecision.State)
	}

	if completedDecision.Votes["A"] != 2 {
		t.Errorf("Expected 2 votes for A, got %d", completedDecision.Votes["A"])
	}
}

func TestEndGame(t *testing.T) {
	service, _ := setupTestServices(t)

	// Create and complete a game
	_, err := service.CreateGame("test-game", "Test Game", 2, 10)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	err = service.EndGame("test-game")
	if err != nil {
		t.Fatalf("Failed to end game: %v", err)
	}

	// Check game status
	status, err := service.GetGameStatus("test-game")
	if err != nil {
		t.Fatalf("Failed to get game status: %v", err)
	}

	if status.Game.State != models.GameStateCompleted {
		t.Errorf("Expected game state 'completed', got '%s'", status.Game.State)
	}

	if !status.Game.IsComplete() {
		t.Error("Expected game to be complete")
	}

	if status.Game.CompletedAt == nil || status.Game.CompletedAt.IsZero() {
		t.Error("Expected completion timestamp to be set")
	}
}

func TestStrategicVoting(t *testing.T) {
	service, _ := setupTestServices(t)
	enhancedVoting := voting.NewEnhancedVotingService()
	enhancedVoting.InitializeStrategies()

	// Create game and decision
	_, err := service.CreateGame("test-game", "Test Game", 2, 10)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	options := []string{"A", "B", "C"}
	_, err = service.StartDecision("test-game", "decision-1", "Test decision", options)
	if err != nil {
		t.Fatalf("Failed to start decision: %v", err)
	}

	// Get game status
	status, err := service.GetGameStatus("test-game")
	if err != nil {
		t.Fatalf("Failed to get game status: %v", err)
	}

	// Cast strategic vote
	err = enhancedVoting.CastStrategicVote(status.Game, status.CurrentDecision, "agent1", "consensus")
	if err != nil {
		t.Fatalf("Failed to cast strategic vote: %v", err)
	}

	// Verify vote was cast by checking the decision directly
	// (since the enhanced voting modifies the decision in place)
	totalVotes := 0
	for _, count := range status.CurrentDecision.Votes {
		totalVotes += count
	}

	if totalVotes != 1 {
		t.Errorf("Expected 1 total vote, got %d", totalVotes)
	}
}
