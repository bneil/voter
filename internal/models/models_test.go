package models_test

import (
	"testing"

	"voter/internal/models"
)

func TestNewGame(t *testing.T) {
	game := models.NewGame("test-game", "Test Game", 3, 10)

	if game.ID != "test-game" {
		t.Errorf("Expected ID 'test-game', got '%s'", game.ID)
	}

	if game.Name != "Test Game" {
		t.Errorf("Expected name 'Test Game', got '%s'", game.Name)
	}

	if game.K != 3 {
		t.Errorf("Expected K=3, got %d", game.K)
	}

	if game.MaxTurns != 10 {
		t.Errorf("Expected MaxTurns=10, got %d", game.MaxTurns)
	}

	if game.State != models.GameStateActive {
		t.Errorf("Expected state 'active', got '%s'", game.State)
	}

	if game.CurrentTurn != 0 {
		t.Errorf("Expected CurrentTurn=0, got %d", game.CurrentTurn)
	}
}

func TestNewDecision(t *testing.T) {
	options := []string{"option1", "option2", "option3"}
	decision := models.NewDecision("test-decision", "game-1", "Test decision", 1, options)

	if decision.ID != "test-decision" {
		t.Errorf("Expected ID 'test-decision', got '%s'", decision.ID)
	}

	if decision.GameID != "game-1" {
		t.Errorf("Expected GameID 'game-1', got '%s'", decision.GameID)
	}

	if decision.Description != "Test decision" {
		t.Errorf("Expected description 'Test decision', got '%s'", decision.Description)
	}

	if decision.TurnNumber != 1 {
		t.Errorf("Expected TurnNumber=1, got %d", decision.TurnNumber)
	}

	if len(decision.Options) != 3 {
		t.Errorf("Expected 3 options, got %d", len(decision.Options))
	}

	if decision.State != models.DecisionStateVoting {
		t.Errorf("Expected state 'voting', got '%s'", decision.State)
	}

	// Check that votes map is initialized
	for _, option := range options {
		if count, exists := decision.Votes[option]; !exists || count != 0 {
			t.Errorf("Expected vote count 0 for option '%s', got %d", option, count)
		}
	}
}

func TestGameIsComplete(t *testing.T) {
	game := models.NewGame("test", "Test", 3, 10)

	// Active game should not be complete
	if game.IsComplete() {
		t.Error("Expected active game to not be complete")
	}

	// Completed game should be complete
	game.State = models.GameStateCompleted
	if !game.IsComplete() {
		t.Error("Expected completed game to be complete")
	}

	// Cancelled game should be complete
	game.State = models.GameStateCancelled
	if !game.IsComplete() {
		t.Error("Expected cancelled game to be complete")
	}
}

func TestGameCanAcceptVotes(t *testing.T) {
	game := models.NewGame("test", "Test", 3, 10)

	// Active game should accept votes
	if !game.CanAcceptVotes() {
		t.Error("Expected active game to accept votes")
	}

	// Paused game should not accept votes
	game.State = models.GameStatePaused
	if game.CanAcceptVotes() {
		t.Error("Expected paused game to not accept votes")
	}

	// Completed game should not accept votes
	game.State = models.GameStateCompleted
	if game.CanAcceptVotes() {
		t.Error("Expected completed game to not accept votes")
	}
}

func TestDecisionCheckWinner(t *testing.T) {
	options := []string{"A", "B", "C"}
	decision := models.NewDecision("test", "game", "test", 1, options)

	// No votes yet
	if winner := decision.CheckWinner(2); winner != nil {
		t.Errorf("Expected no winner with no votes, got %s", *winner)
	}

	// Add votes: A=3, B=1, C=0
	decision.AddVote("A")
	decision.AddVote("A")
	decision.AddVote("A")
	decision.AddVote("B")

	// A should be winner (3-1=2 >= K=2)
	if winner := decision.CheckWinner(2); winner == nil || *winner != "A" {
		t.Errorf("Expected A to be winner, got %v", winner)
	}

	// With K=3, A should not be winner (3-1=2 < 3)
	if winner := decision.CheckWinner(3); winner != nil {
		t.Errorf("Expected no winner with K=3, got %s", *winner)
	}
}

func TestDecisionAddVote(t *testing.T) {
	options := []string{"A", "B", "C"}
	decision := models.NewDecision("test", "game", "test", 1, options)

	// Valid vote
	if !decision.AddVote("A") {
		t.Error("Expected valid vote to succeed")
	}

	if decision.Votes["A"] != 1 {
		t.Errorf("Expected vote count 1 for A, got %d", decision.Votes["A"])
	}

	// Invalid option
	if decision.AddVote("D") {
		t.Error("Expected invalid vote to fail")
	}

	// Close voting
	decision.State = models.DecisionStateCompleted

	// Vote after closing should fail
	if decision.AddVote("B") {
		t.Error("Expected vote after closing to fail")
	}
}

func TestGetCurrentDecision(t *testing.T) {
	game := models.NewGame("test", "Test", 3, 10)

	// No decisions yet
	if decision := game.GetCurrentDecision(); decision != nil {
		t.Error("Expected no current decision for new game")
	}

	// Add a completed decision
	completedDecision := models.NewDecision("completed", "test", "completed", 1, []string{"A", "B"})
	completedDecision.State = models.DecisionStateCompleted
	game.Decisions = append(game.Decisions, *completedDecision)

	// Still no current decision
	if decision := game.GetCurrentDecision(); decision != nil {
		t.Error("Expected no current decision when only completed decisions exist")
	}

	// Add an active decision
	activeDecision := models.NewDecision("active", "test", "active", 2, []string{"A", "B"})
	game.Decisions = append(game.Decisions, *activeDecision)

	// Should return the active decision
	if decision := game.GetCurrentDecision(); decision == nil || decision.ID != "active" {
		t.Errorf("Expected active decision, got %v", decision)
	}
}
