package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"voter/internal/metrics"
	"voter/internal/models"
	"voter/internal/project"
	"voter/internal/storage"
	"voter/internal/voting"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Initialize dependencies
	store, err := storage.NewJSONProjectStore("./data")
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	votingService := project.NewVotingService()
	projectService := project.NewService(store, votingService)
	enhancedVoting := voting.NewEnhancedVotingService()
	enhancedVoting.InitializeStrategies()
	scorer := metrics.NewScorer()
	metricsTracker := metrics.NewTracker()

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "create-project":
		handleCreateProject(projectService, args)
	case "start-decision":
		handleStartDecision(projectService, args)
	case "vote":
		handleVote(projectService, args)
	case "close-voting":
		handleCloseVoting(projectService, args)
	case "project-status":
		handleProjectStatus(projectService, scorer, args)
	case "list-projects":
		handleListProjects(projectService, args)
	case "project-stats":
		handleProjectStats(metricsTracker, args)
	case "simulate-voting":
		handleSimulateVoting(projectService, enhancedVoting, args)
	case "strategic-vote":
		handleStrategicVote(projectService, enhancedVoting, args)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleCreateProject(service *project.Service, args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: create-game <id> <name> <k> <max-turns>")
		os.Exit(1)
	}

	id := args[0]
	name := args[1]
	k, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Printf("Invalid K value: %v\n", err)
		os.Exit(1)
	}

	maxTurns := 10 // default
	if len(args) > 3 {
		maxTurns, err = strconv.Atoi(args[3])
		if err != nil {
			fmt.Printf("Invalid max-turns value: %v\n", err)
			os.Exit(1)
		}
	}

	project, err := service.CreateProject(id, name, k, maxTurns)
	if err != nil {
		fmt.Printf("Failed to create project: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Project created successfully:\n")
	printProject(project)
}

func handleStartDecision(service *project.Service, args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: start-decision <project-id> <description> <option1> <option2> [option3...]")
		os.Exit(1)
	}

	projectID := args[0]
	description := args[1]
	options := args[2:]

	if len(options) < 2 {
		fmt.Println("At least 2 options required")
		os.Exit(1)
	}

	decision, err := service.StartDecision(projectID, fmt.Sprintf("decision_%d", 1), description, options)
	if err != nil {
		fmt.Printf("Failed to start decision: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Decision started:\n")
	printDecision(decision)
}

func handleVote(service *project.Service, args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: vote <project-id> <decision-id> <agent-id> <option>")
		os.Exit(1)
	}

	projectID := args[0]
	decisionID := args[1]
	agentID := args[2]
	option := args[3]

	err := service.CastVote(projectID, decisionID, agentID, option)
	if err != nil {
		fmt.Printf("Failed to cast vote: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Vote cast successfully by agent %s for option '%s'\n", agentID, option)
}

func handleCloseVoting(service *project.Service, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: close-voting <project-id>")
		os.Exit(1)
	}

	projectID := args[0]

	err := service.EndProject(projectID)
	if err != nil {
		fmt.Printf("Failed to close voting: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Voting closed for project %s\n", projectID)
}

func handleProjectStatus(service *project.Service, scorer *metrics.Scorer, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: project-status <project-id>")
		os.Exit(1)
	}

	projectID := args[0]

	status, err := service.GetProjectStatus(projectID)
	if err != nil {
		fmt.Printf("Failed to get project status: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Project Status:\n")
	fmt.Printf("ID: %s\n", status.Project.ID)
	fmt.Printf("Name: %s\n", status.Project.Name)
	fmt.Printf("State: %s\n", status.Project.State)
	fmt.Printf("Current Turn: %d/%d\n", status.Project.CurrentTurn, status.Project.MaxTurns)
	fmt.Printf("Active: %t\n", status.IsActive)

	if status.CurrentDecision != nil {
		fmt.Printf("\nCurrent Decision:\n")
		fmt.Printf("ID: %s\n", status.CurrentDecision.ID)
		fmt.Printf("Description: %s\n", status.CurrentDecision.Description)
		fmt.Printf("State: %s\n", status.CurrentDecision.State)
		fmt.Printf("Options: %s\n", strings.Join(status.CurrentDecision.Options, ", "))

		if len(status.VoteCounts) > 0 {
			fmt.Printf("Vote Counts:\n")
			for option, count := range status.VoteCounts {
				fmt.Printf("  %s: %d\n", option, count)
			}
		}
	}

	// Show score if project is complete
	if status.Project.IsComplete() {
		score := scorer.CalculateProjectScore(status.Project)
		if score != nil {
			fmt.Printf("\nFinal Score: %d\n", score.TotalScore)
			fmt.Printf("Completion Bonus: %d\n", score.CompletionBonus)
			fmt.Printf("Efficiency Bonus: %d\n", score.EfficiencyBonus)
			fmt.Printf("Participation Bonus: %d\n", score.ParticipationBonus)
		}
	}
}

func handleListProjects(service *project.Service, args []string) {
	projects, err := service.ListProjects()
	if err != nil {
		fmt.Printf("Failed to list projects: %v\n", err)
		os.Exit(1)
	}

	if len(projects) == 0 {
		fmt.Println("No projects found")
		return
	}

	fmt.Printf("Projects:\n")
	for _, project := range projects {
		fmt.Printf("- %s: %s (%s) - Turn %d/%d\n",
			project.ID, project.Name, project.State, project.CurrentTurn, project.MaxTurns)
	}
}

func handleSimulateVoting(service *project.Service, enhancedVoting *voting.EnhancedVotingService, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: simulate-voting <project-id> <decision-id> <agent-count>")
		os.Exit(1)
	}

	projectID := args[0]
	decisionID := args[1]
	agentCount, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Printf("Invalid agent count: %v\n", err)
		os.Exit(1)
	}

	// Get the current decision
	status, err := service.GetProjectStatus(projectID)
	if err != nil {
		fmt.Printf("Failed to get project status: %v\n", err)
		os.Exit(1)
	}

	if status.CurrentDecision == nil || status.CurrentDecision.ID != decisionID {
		fmt.Println("Decision not found or not active")
		os.Exit(1)
	}

	err = enhancedVoting.SimulateAgentVoting(status.Project, status.CurrentDecision, agentCount)
	if err != nil {
		fmt.Printf("Failed to simulate voting: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Simulated %d agents voting\n", agentCount)
}

func handleProjectStats(tracker *metrics.Tracker, args []string) {
	stats := tracker.GetGlobalStats()

	fmt.Printf("Global Statistics:\n")
	fmt.Printf("Total Projects: %d\n", stats.TotalProjects)
	fmt.Printf("Total Decisions: %d\n", stats.TotalDecisions)
	fmt.Printf("Average Project Score: %.1f\n", stats.AverageProjectScore)
	fmt.Printf("Average Consensus Time: %v\n", stats.AverageConsensusTime)
	if stats.BestProjectID != "" {
		fmt.Printf("Best Project: %s (Score: %d)\n", stats.BestProjectID, stats.BestProjectScore)
	}
}

func handleStrategicVote(service *project.Service, enhancedVoting *voting.EnhancedVotingService, args []string) {
	if len(args) < 4 {
		fmt.Println("Usage: strategic-vote <project-id> <decision-id> <agent-id> <strategy>")
		os.Exit(1)
	}

	projectID := args[0]
	decisionID := args[1]
	agentID := args[2]
	strategy := args[3]

	// Get the current decision
	status, err := service.GetProjectStatus(projectID)
	if err != nil {
		fmt.Printf("Failed to get project status: %v\n", err)
		os.Exit(1)
	}

	if status.CurrentDecision == nil || status.CurrentDecision.ID != decisionID {
		fmt.Println("Decision not found or not active")
		os.Exit(1)
	}

	err = enhancedVoting.CastStrategicVote(status.Project, status.CurrentDecision, agentID, strategy)
	if err != nil {
		fmt.Printf("Failed to cast strategic vote: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Strategic vote cast by agent %s using %s strategy\n", agentID, strategy)
}

func printProject(project *models.Project) {
	data, _ := json.MarshalIndent(project, "", "  ")
	fmt.Println(string(data))
}

func printDecision(decision *models.Decision) {
	data, _ := json.MarshalIndent(decision, "", "  ")
	fmt.Println(string(data))
}

func printUsage() {
	fmt.Println("Voter - First-to-Ahead-by-K Voting System")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create-project <id> <name> <k> [max-turns]    Create a new project")
	fmt.Println("  start-decision <project-id> <desc> <opt1> <opt2> [opt3...]  Start a voting decision")
	fmt.Println("  vote <project-id> <decision-id> <agent-id> <option>          Cast a vote")
	fmt.Println("  strategic-vote <project-id> <decision-id> <agent-id> <strategy>  Cast strategic vote")
	fmt.Println("  simulate-voting <project-id> <decision-id> <agent-count>     Simulate agent voting")
	fmt.Println("  close-voting <project-id>                          Close voting for project")
	fmt.Println("  project-status <project-id>                           Show project status")
	fmt.Println("  list-projects                                  List all projects")
	fmt.Println("  project-stats                                  Show global statistics")
	fmt.Println()
	fmt.Println("Strategies: random, consensus, optimal")
}
