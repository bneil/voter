package storage

import "voter/internal/models"

// GameStore defines the interface for game storage operations
type GameStore interface {
	SaveGame(game *models.Game) error
	GetGame(id string) (*models.Game, error)
	ListGames() ([]*models.Game, error)
	DeleteGame(id string) error
}

// VoteStore defines the interface for vote storage operations
type VoteStore interface {
	SaveVote(vote *models.Vote) error
	GetVotesByDecision(decisionID string) ([]*models.Vote, error)
	GetVotesByGame(gameID string) ([]*models.Vote, error)
}
