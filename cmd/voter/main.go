package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"voter/internal/game"
	"voter/internal/metrics"
	"voter/internal/models"
	"voter/internal/storage"
	"voter/internal/voting"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Initialize dependencies
	store, err := storage.NewJSONGameStore("./data")
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	votingService := game.NewVotingService()
	gameService := game.NewService(store, votingService)
	enhancedVoting := voting.NewEnhancedVotingService()
	enhancedVoting.InitializeStrategies()
	scorer := metrics.NewScorer()
	metricsTracker := metrics.NewTracker()

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "create-project":
		handleCreateProject(gameService, args)
	case "start-decision":
		handleStartDecision(gameService, args)
	case "vote":
		handleVote(gameService, args)
	case "close-voting":
		handleCloseVoting(gameService, args)
	case "project-status":
		handleProjectStatus(gameService, scorer, args)
	case "list-projects":
		handleListProjects(gameService, args)
	case "project-stats":
		handleProjectStats(metricsTracker, args)
	case "simulate-voting":
		handleSimulateVoting(gameService, enhancedVoting, args)
	case "strategic-vote":
		handleStrategicVote(gameService, enhancedVoting, args)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleCreateProject(service *game.Service, args []string) {
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

	game, err := service.CreateGame(id, name, k, maxTurns)
	if err != nil {
		fmt.Printf("Failed to create game: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Game created successfully:\n")
	printProject(game)
}

func handleStartDecision(service *game.Service, args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: start-decision <game-id> <description> <option1> <option2> [option3...]")
		os.Exit(1)
	}

	gameID := args[0]
	description := args[1]
	options := args[2:]

	if len(options) < 2 {
		fmt.Println("At least 2 options required")
		os.Exit(1)
	}

	decision, err := service.StartDecision(gameID, fmt.Sprintf("decision_%d", 1), description, options)
	if err != nil {
		fmt.Printf("Failed to start decision: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Decision started:\n")
	printDecision(decision)
}

func handleVote(service *game.Service, args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: vote <game-id> <decision-id> <agent-id> <option>")
		os.Exit(1)
	}

	gameID := args[0]
	decisionID := args[1]
	agentID := args[2]
	option := args[3]

	err := service.CastVote(gameID, decisionID, agentID, option)
	if err != nil {
		fmt.Printf("Failed to cast vote: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Vote cast successfully by agent %s for option '%s'\n", agentID, option)
}

func handleCloseVoting(service *game.Service, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: close-voting <game-id>")
		os.Exit(1)
	}

	gameID := args[0]

	err := service.EndGame(gameID)
	if err != nil {
		fmt.Printf("Failed to close voting: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Voting closed for game %s\n", gameID)
}

func handleProjectStatus(service *game.Service, scorer *metrics.Scorer, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: game-status <game-id>")
		os.Exit(1)
	}

	gameID := args[0]

	status, err := service.GetGameStatus(gameID)
	if err != nil {
		fmt.Printf("Failed to get game status: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Game Status:\n")
	fmt.Printf("ID: %s\n", status.Game.ID)
	fmt.Printf("Name: %s\n", status.Game.Name)
	fmt.Printf("State: %s\n", status.Game.State)
	fmt.Printf("Current Turn: %d/%d\n", status.Game.CurrentTurn, status.Game.MaxTurns)
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

	// Show score if game is complete
	if status.Game.IsComplete() {
		score := scorer.CalculateGameScore(status.Game)
		if score != nil {
			fmt.Printf("\nFinal Score: %d\n", score.TotalScore)
			fmt.Printf("Completion Bonus: %d\n", score.CompletionBonus)
			fmt.Printf("Efficiency Bonus: %d\n", score.EfficiencyBonus)
			fmt.Printf("Participation Bonus: %d\n", score.ParticipationBonus)
		}
	}
}

func handleListProjects(service *game.Service, args []string) {
	games, err := service.ListGames()
	if err != nil {
		fmt.Printf("Failed to list games: %v\n", err)
		os.Exit(1)
	}

	if len(games) == 0 {
		fmt.Println("No games found")
		return
	}

	fmt.Printf("Games:\n")
	for _, game := range games {
		fmt.Printf("- %s: %s (%s) - Turn %d/%d\n",
			game.ID, game.Name, game.State, game.CurrentTurn, game.MaxTurns)
	}
}

func handleSimulateVoting(service *game.Service, enhancedVoting *voting.EnhancedVotingService, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: simulate-voting <game-id> <decision-id> <agent-count>")
		os.Exit(1)
	}

	gameID := args[0]
	decisionID := args[1]
	agentCount, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Printf("Invalid agent count: %v\n", err)
		os.Exit(1)
	}

	// Get the current decision
	status, err := service.GetGameStatus(gameID)
	if err != nil {
		fmt.Printf("Failed to get game status: %v\n", err)
		os.Exit(1)
	}

	if status.CurrentDecision == nil || status.CurrentDecision.ID != decisionID {
		fmt.Println("Decision not found or not active")
		os.Exit(1)
	}

	err = enhancedVoting.SimulateAgentVoting(status.Game, status.CurrentDecision, agentCount)
	if err != nil {
		fmt.Printf("Failed to simulate voting: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Simulated %d agents voting\n", agentCount)
}

func handleProjectStats(tracker *metrics.Tracker, args []string) {
	stats := tracker.GetGlobalStats()

	fmt.Printf("Global Statistics:\n")
	fmt.Printf("Total Games: %d\n", stats.TotalGames)
	fmt.Printf("Total Decisions: %d\n", stats.TotalDecisions)
	fmt.Printf("Average Game Score: %.1f\n", stats.AverageGameScore)
	fmt.Printf("Average Consensus Time: %v\n", stats.AverageConsensusTime)
	if stats.BestGameID != "" {
		fmt.Printf("Best Game: %s (Score: %d)\n", stats.BestGameID, stats.BestGameScore)
	}
}

func handleStrategicVote(service *game.Service, enhancedVoting *voting.EnhancedVotingService, args []string) {
	if len(args) < 4 {
		fmt.Println("Usage: strategic-vote <game-id> <decision-id> <agent-id> <strategy>")
		os.Exit(1)
	}

	gameID := args[0]
	decisionID := args[1]
	agentID := args[2]
	strategy := args[3]

	// Get the current decision
	status, err := service.GetGameStatus(gameID)
	if err != nil {
		fmt.Printf("Failed to get game status: %v\n", err)
		os.Exit(1)
	}

	if status.CurrentDecision == nil || status.CurrentDecision.ID != decisionID {
		fmt.Println("Decision not found or not active")
		os.Exit(1)
	}

	err = enhancedVoting.CastStrategicVote(status.Game, status.CurrentDecision, agentID, strategy)
	if err != nil {
		fmt.Printf("Failed to cast strategic vote: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Strategic vote cast by agent %s using %s strategy\n", agentID, strategy)
}

func printProject(project *models.Game) {
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
