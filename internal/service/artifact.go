package service

import (
	"context"
	"errors"
	"io"

	"github.com/istvzsig/artifix/internal/artifact"
)

type ArtifactRepository interface {
	// Create inserts a new artifact and rejects duplicate coordinates.
	Create(ctx context.Context, a artifact.Artifact) error
	// Update persists changes to an existing artifact's lifecycle.
	Update(ctx context.Context, a artifact.Artifact) error
}

type BinaryStorage interface {
	// Stage streams content into temporary storage and calculates its SHA-256 digest and size.
	Stage(ctx context.Context, content io.Reader) (artifact.Digest, int64, error)
	// Publish moves staged content into permanent content-addressed storage.
	Publish(ctx context.Context, digest artifact.Digest) error
	// Delete removes staged or published content when appropriate.
	Delete(ctx context.Context, digest artifact.Digest) error
}

type Service struct {
	repository ArtifactRepository
	storage    BinaryStorage
}

func NewService(repository ArtifactRepository, storage BinaryStorage) (*Service, error) {
	if repository == nil {
		return nil, errors.New("artifact repository is required")
	}

	if storage == nil {
		return nil, errors.New("binary storage is required")
	}

	return &Service{
		repository: repository,
		storage:    storage,
	}, nil
}
