package service

import (
	"context"
	"io"
	"time"

	"github.com/istvzsig/artifix/internal/artifact"
	"github.com/istvzsig/artifix/internal/repository"
	"github.com/istvzsig/artifix/internal/storage"
)

type Service struct {
	repository repository.ArtifactRepository
	storage    storage.BinaryStorage
}

func NewService(
	repository repository.ArtifactRepository,
	storage storage.BinaryStorage,
) (*Service, error) {
	if repository == nil {
		return nil, ErrRepositoryRequired
	}

	if storage == nil {
		return nil, ErrBinaryStorageRequired
	}

	return &Service{
		repository: repository,
		storage:    storage,
	}, nil
}

func (s *Service) Upload(
	ctx context.Context,
	coords artifact.Coordinates,
	content io.Reader,
	mediaType string,
) (artifact.Artifact, error) {
	digest, size, err := s.storage.Stage(ctx, content)
	if err != nil {
		return artifact.Artifact{}, err
	}

	a, err := artifact.NewArtifact(coords, digest, size, mediaType)
	if err != nil {
		if cleanupErr := s.cleanupStagedContent(ctx, digest); cleanupErr != nil {
			// Log cleanupErr later.
		}
		return artifact.Artifact{}, err
	}

	if err := s.repository.Create(ctx, a); err != nil {
		if cleanupErr := s.cleanupStagedContent(ctx, digest); cleanupErr != nil {
			// Log cleanupErr later.
		}
		return artifact.Artifact{}, err
	}

	if err := s.storage.Publish(ctx, digest); err != nil {
		if cleanupErr := s.cleanupStagedContent(ctx, digest); cleanupErr != nil {
			// Log cleanupErr later.
		}
		return artifact.Artifact{}, err
	}

	if err := a.MarkAvailable(); err != nil {
		return artifact.Artifact{}, err
	}

	if err := s.repository.Update(ctx, a); err != nil {
		return artifact.Artifact{}, err
	}

	return a, nil
}

func (s *Service) Delete(
	ctx context.Context,
	coords artifact.Coordinates,
) error {
	return s.repository.Delete(ctx, coords)
}

const cleanupTimeout = 5 * time.Second

func (s *Service) cleanupStagedContent(
	ctx context.Context,
	digest artifact.Digest,
) error {
	cleanupCtx, cancel := context.WithTimeout(
		context.WithoutCancel(ctx),
		cleanupTimeout,
	)
	defer cancel()

	return s.storage.DiscardStage(cleanupCtx, digest)
}
