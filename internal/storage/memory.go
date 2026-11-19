package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"sync"

	"github.com/istvzsig/artifix/internal/artifact"
)

type MemoryStorage struct {
	mu        sync.RWMutex
	temporary map[artifact.Digest][]byte
	published map[artifact.Digest][]byte
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		temporary: make(map[artifact.Digest][]byte),
		published: make(map[artifact.Digest][]byte),
	}
}

var _ BinaryStorage = (*MemoryStorage)(nil)

const maxArtifactSize = 100 << 20 // 100 MiB

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

// Important: this checks cancellation between reads.
// It cannot interrupt an underlying reader that's blocked inside Read.
// Also, this implementation is appropriate for an in-memory adapter;
// a filesystem or S3 implementation should stream to its destination
// rather than buffer the entire artifact in RAM.
func (r *contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}

	return r.reader.Read(p)
}

func (s *MemoryStorage) Stage(
	ctx context.Context,
	content io.Reader,
) (artifact.Digest, int64, error) {
	if err := ctx.Err(); err != nil {
		return artifact.Digest{}, 0, err
	}

	buf := bytes.Buffer{}
	hash := sha256.New()

	writer := io.MultiWriter(&buf, hash)
	limited := io.LimitReader(content, maxArtifactSize+1)

	reader := &contextReader{
		ctx:    ctx,
		reader: limited,
	}

	size, err := io.Copy(writer, reader)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return artifact.Digest{}, 0, ctxErr
		}

		return artifact.Digest{}, 0, fmt.Errorf(
			"%w: %w",
			ErrFailedToReadContent,
			err,
		)
	}

	if size > maxArtifactSize {
		return artifact.Digest{}, 0, ErrArtifactTooLarge
	}

	if err := ctx.Err(); err != nil {
		return artifact.Digest{}, 0, err
	}

	digest := artifact.Digest(hash.Sum(nil))

	s.mu.Lock()
	s.temporary[digest] = bytes.Clone(buf.Bytes())
	s.mu.Unlock()

	return digest, size, nil
}

func (s *MemoryStorage) Publish(ctx context.Context, digest artifact.Digest) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.published[digest]
	if ok {
		return nil
	}

	staged, ok := s.temporary[digest]
	if !ok {
		return ErrStagedContentNotFound
	}

	delete(s.temporary, digest)
	s.published[digest] = staged

	return nil
}

func (s *MemoryStorage) DiscardStage(ctx context.Context, digest artifact.Digest) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return nil
}
