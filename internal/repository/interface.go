package repository

import (
	"context"

	"github.com/istvzsig/artifix/internal/artifact"
)

type ArtifactRepository interface {
	Create(ctx context.Context, a artifact.Artifact) error
	Get(ctx context.Context, coords artifact.Coordinates) (artifact.Artifact, error)
	Update(ctx context.Context, a artifact.Artifact) error
}
