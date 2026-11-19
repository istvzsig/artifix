package storage

import (
	"context"
	"io"

	"github.com/istvzsig/artifix/internal/artifact"
)

type BinaryStorage interface {
	Stage(ctx context.Context, content io.Reader) (artifact.Digest, int64, error)
	Publish(ctx context.Context, digest artifact.Digest) error
	DiscardStage(ctx context.Context, digest artifact.Digest) error
}
