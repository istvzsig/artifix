package repository

import (
	"context"
	"sync"

	"github.com/istvzsig/artifix/internal/artifact"
)

type MemoryRepository struct {
	mu        sync.RWMutex
	artifacts map[artifact.Coordinates]artifact.Artifact
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		artifacts: make(map[artifact.Coordinates]artifact.Artifact),
	}
}

var _ ArtifactRepository = (*MemoryRepository)(nil)

func (r *MemoryRepository) Create(
	ctx context.Context,
	a artifact.Artifact,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.artifacts[a.Coordinates()]; ok {
		return ErrArtifactAlreadyExists
	}

	r.artifacts[a.Coordinates()] = a

	return nil
}

func (r *MemoryRepository) Get(
	ctx context.Context,
	coords artifact.Coordinates,
) (artifact.Artifact, error) {
	if err := ctx.Err(); err != nil {
		return artifact.Artifact{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	a, ok := r.artifacts[coords]
	if !ok {
		return artifact.Artifact{}, ErrArtifactNotFound
	}

	return a, nil
}

func (r *MemoryRepository) Update(
	ctx context.Context,
	a artifact.Artifact,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	exists, ok := r.artifacts[a.Coordinates()]
	if !ok {
		return ErrArtifactNotFound
	}

	if exists.ID() != a.ID() {
		return ErrArtifactIDMismatch
	}

	r.artifacts[a.Coordinates()] = a

	return nil
}
