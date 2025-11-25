package project_test

import (
	"testing"

	"github.com/bneil/voter/internal/models"
	"github.com/bneil/voter/internal/project"
	"github.com/bneil/voter/internal/storage"
	"github.com/bneil/voter/internal/voting"
)

func setupTestServices(t *testing.T) (*project.Service, *storage.JSONProjectStore) {
	t.Helper()

	store, err := storage.NewJSONProjectStore(t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}

	votingService := project.NewVotingService()
	service := project.NewService(store, votingService)

	return service, store
}

func TestCreateProject(t *testing.T) {
	service, _ := setupTestServices(t)

	project, err := service.CreateProject("test-project", "Test Project", 3, 10)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	if project.ID != "test-project" {
		t.Errorf("Expected ID 'test-project', got '%s'", project.ID)
	}

	if project.K != 3 {
		t.Errorf("Expected K=3, got %d", project.K)
	}

	if project.MaxTurns != 10 {
		t.Errorf("Expected MaxTurns=10, got %d", project.MaxTurns)
	}
}

func TestStartDecision(t *testing.T) {
	service, _ := setupTestServices(t)

	// Create project
	_, err := service.CreateProject("test-project", "Test Project", 3, 10)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Start decision
	options := []string{"option1", "option2", "option3"}
	decision, err := service.StartDecision("test-project", "decision-1", "Test decision", options)
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

	// Create project with K=2
	_, err := service.CreateProject("test-project", "Test Project", 2, 10)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	options := []string{"A", "B", "C"}
	decision, err := service.StartDecision("test-project", "decision-1", "Test decision", options)
	if err != nil {
		t.Fatalf("Failed to start decision: %v", err)
	}

	// Cast votes
	err = service.CastVote("test-project", decision.ID, "agent1", "A")
	if err != nil {
		t.Fatalf("Failed to cast vote: %v", err)
	}

	err = service.CastVote("test-project", decision.ID, "agent2", "A")
	if err != nil {
		t.Fatalf("Failed to cast vote: %v", err)
	}

	// Check that decision completed with A as winner
	status, err := service.GetProjectStatus("test-project")
	if err != nil {
		t.Fatalf("Failed to get project status: %v", err)
	}

	if len(status.Project.Decisions) == 0 {
		t.Fatal("Expected at least one decision in project")
	}

	completedDecision := status.Project.Decisions[len(status.Project.Decisions)-1]
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

func TestEndProject(t *testing.T) {
	service, _ := setupTestServices(t)

	// Create and complete a project
	_, err := service.CreateProject("test-project", "Test Project", 2, 10)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	err = service.EndProject("test-project")
	if err != nil {
		t.Fatalf("Failed to end project: %v", err)
	}

	// Check project status
	status, err := service.GetProjectStatus("test-project")
	if err != nil {
		t.Fatalf("Failed to get project status: %v", err)
	}

	if status.Project.State != models.ProjectStateCompleted {
		t.Errorf("Expected project state 'completed', got '%s'", status.Project.State)
	}

	if !status.Project.IsComplete() {
		t.Error("Expected game to be complete")
	}

	if status.Project.CompletedAt == nil || status.Project.CompletedAt.IsZero() {
		t.Error("Expected completion timestamp to be set")
	}
}

func TestStrategicVoting(t *testing.T) {
	service, _ := setupTestServices(t)
	enhancedVoting := voting.NewEnhancedVotingService()
	enhancedVoting.InitializeStrategies()

	// Create project and decision
	_, err := service.CreateProject("test-project", "Test Project", 2, 10)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	options := []string{"A", "B", "C"}
	_, err = service.StartDecision("test-project", "decision-1", "Test decision", options)
	if err != nil {
		t.Fatalf("Failed to start decision: %v", err)
	}

	// Get project status
	status, err := service.GetProjectStatus("test-project")
	if err != nil {
		t.Fatalf("Failed to get project status: %v", err)
	}

	// Cast strategic vote
	err = enhancedVoting.CastStrategicVote(status.Project, status.CurrentDecision, "agent1", "consensus")
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
