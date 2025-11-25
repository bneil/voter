package storage

import "voter/internal/models"

// ProjectStore defines the interface for project storage operations
type ProjectStore interface {
	SaveProject(project *models.Project) error
	GetProject(id string) (*models.Project, error)
	ListProjects() ([]*models.Project, error)
	DeleteProject(id string) error
}

// VoteStore defines the interface for vote storage operations
type VoteStore interface {
	SaveVote(vote *models.Vote) error
	GetVotesByDecision(decisionID string) ([]*models.Vote, error)
	GetVotesByProject(projectID string) ([]*models.Vote, error)
}
