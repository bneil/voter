# Voter

First-to-Ahead-by-K voting system for rapid consensus formation.

## Quick Start

```bash
make build
./bin/voter create-project "tower-hanoi" "Tower of Hanoi Solver" 3 10
./bin/voter start-decision "tower-hanoi" "Move disk from A to B" "A->B" "A->C" "B->A" "B->C" "C->A" "C->B"
./bin/voter strategic-vote "tower-hanoi" "decision_1" "agent1" "consensus"
./bin/voter project-status "tower-hanoi"
```

## Commands

- `create-project <name> <desc> <k> <agents>` - Create voting project
- `start-decision <project> <desc> <options...>` - Start decision with options
- `vote <project> <decision> <agent> <option>` - Cast vote
- `strategic-vote <project> <decision> <agent> <strategy>` - Strategic voting
- `simulate-voting <project> <decision> <agents>` - Simulate multiple agents
- `project-status <project>` - Show project status
- `list-projects` - List all projects
- `project-stats` - Show statistics

## Strategies

- `random` - Random selection
- `consensus` - Follow momentum
- `optimal` - Game-specific optimal

## Testing

```bash
make test
make test-coverage
```
