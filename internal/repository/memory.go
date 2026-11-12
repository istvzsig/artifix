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

func (r *MemoryRepository) List(
	ctx context.Context,
	namespace string,
	name string,
) ([]artifact.Artifact, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := []artifact.Artifact{}

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, a := range r.artifacts {
		if a.Coordinates().Namespace() == namespace &&
			a.Coordinates().Name() == name {
			result = append(result, a)
		}
	}

	return result, nil
}

func (r *MemoryRepository) Delete(
	ctx context.Context,
	coords artifact.Coordinates,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	a, ok := r.artifacts[coords]
	if !ok {
		return ErrArtifactNotFound
	}

	if err := a.Delete(); err != nil {
		return err
	}

	r.artifacts[coords] = a

	return nil
}
