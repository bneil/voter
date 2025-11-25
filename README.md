# Voter - First-to-Ahead-by-K Voting System

A command-line voting tool that implements First-to-Ahead-by-K voting for rapid consensus formation in decision-making scenarios.

## Features

- **First-to-Ahead-by-K Voting**: A candidate must lead by K votes to win
- **Atomic Operations**: Semaphore-protected voting to prevent race conditions
- **Game Integration**: Supports complex decision scenarios like Tower of Hanoi
- **Strategic Voting**: Multiple voting strategies (random, consensus, optimal)
- **Performance Metrics**: Tracks voting efficiency and consensus speed

## Quick Start

### Build
```bash
go build -o voter ./cmd/voter
```

### Create a Project
```bash
./voter create-project "tower-hanoi" "Tower of Hanoi Solver" 3 10
```

### Start a Decision
```bash
./voter start-decision "tower-hanoi" "Move disk from A to B" "A->B" "A->C" "B->A" "B->C" "C->A" "C->B"
```

### Cast Votes
```bash
./voter vote "tower-hanoi" "decision_1" "agent1" "A->B"
./voter strategic-vote "tower-hanoi" "decision_1" "agent2" "consensus"
```

### Check Status
```bash
./voter project-status "tower-hanoi"
```

### Simulate Multiple Agents
```bash
./voter simulate-voting "tower-hanoi" "decision_1" 10
```

### List Projects
```bash
./voter list-projects
```

### Show Statistics
```bash
./voter project-stats
```

## Voting Strategies

- **random**: Random selection among options
- **consensus**: Votes for options gaining momentum
- **optimal**: Game-specific optimal strategies

## Architecture

- `internal/models/`: Data structures for games, decisions, and votes
- `internal/game/`: Game session management and progression
- `internal/voting/`: Voting strategies and enhanced voting logic
- `internal/metrics/`: Performance scoring and analytics
- `internal/storage/`: JSON-based persistence layer

## Testing

Run all tests:
```bash
go test ./...
```

Run with coverage:
```bash
go test -cover ./...
```

## Releases

Create a new release by tagging a commit:

```bash
git tag v1.0.0
git push origin v1.0.0
```

The GitHub Actions workflow will automatically:
- Build binaries for Linux, macOS (Intel + Apple Silicon), and Windows
- Run tests to ensure code quality
- Create a GitHub release with downloadable binaries

## Game Theory

This system implements First-to-Ahead-by-K voting where:
1. Multiple agents vote in parallel on decision options
2. The first option to achieve K more votes than any other wins
3. This creates rapid local consensus with exponential accuracy gains

Perfect for AI agents making sequential decisions in puzzle-solving or strategic games.
