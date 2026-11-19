package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"strings"
	"testing"

	"github.com/istvzsig/artifix/internal/artifact"
)

func TestMemoryStorage_Stage_Success(t *testing.T) {
	storage := NewMemoryStorage()
	content := "test content"

	digest, size, err := storage.Stage(
		t.Context(),
		strings.NewReader(content),
	)
	if err != nil {
		t.Fatal(err)
	}

	expectedDigest := sha256.Sum256([]byte(content))
	if digest != artifact.Digest(expectedDigest) {
		t.Errorf("expected digest %x, got %x", expectedDigest, digest)
	}

	if size != int64(len(content)) {
		t.Errorf("expected size %d, got %d", len(content), size)
	}

	stored, ok := stagedContent(storage, digest)
	if !ok {
		t.Fatal("expected content to be stored")
	}

	if !bytes.Equal(stored, []byte(content)) {
		t.Errorf("expected stored content %q, got %q", content, stored)
	}
}

func TestMemoryStorage_Stage_EmptyContent(t *testing.T) {
	storage := NewMemoryStorage()

	digest, size, err := storage.Stage(
		t.Context(),
		strings.NewReader(""),
	)
	if err != nil {
		t.Fatal(err)
	}

	if size != 0 {
		t.Errorf("expected size 0, got %d", size)
	}

	expectedDigest := sha256.Sum256(nil)
	if digest != artifact.Digest(expectedDigest) {
		t.Errorf("expected digest %x, got %x", expectedDigest, digest)
	}

	if _, ok := stagedContent(storage, digest); !ok {
		t.Fatal("expected empty content to be staged")
	}
}

func TestMemoryStorage_Stage_TooLarge(t *testing.T) {
	storage := NewMemoryStorage()
	content := strings.Repeat("a", int(maxArtifactSize)+1)

	digest, size, err := storage.Stage(
		t.Context(),
		strings.NewReader(content),
	)

	if !errors.Is(err, ErrArtifactTooLarge) {
		t.Fatalf("expected error %v, got %v", ErrArtifactTooLarge, err)
	}

	if digest != (artifact.Digest{}) {
		t.Errorf("expected zero digest, got %x", digest)
	}

	if size != 0 {
		t.Errorf("expected size 0, got %d", size)
	}

	storage.mu.RLock()
	got := len(storage.temporary)
	storage.mu.RUnlock()

	if got != 0 {
		t.Errorf("expected empty temporary storage, got %d entries", got)
	}
}

type failingReader struct{}

func (f failingReader) Read([]byte) (int, error) {
	return 0, errors.New("simulated read failure")
}

func TestMemoryStorage_Stage_ReadFailure(t *testing.T) {
	storage := NewMemoryStorage()

	_, _, err := storage.Stage(t.Context(), failingReader{})

	if !errors.Is(err, ErrFailedToReadContent) {
		t.Fatalf("expected error %v, got %v", ErrFailedToReadContent, err)
	}

	if len(storage.temporary) != 0 {
		t.Errorf("expected empty temporary storage, got %d entries", len(storage.temporary))
	}
}

func TestMemoryStorage_Stage_CanceledContext(t *testing.T) {
	storage := NewMemoryStorage()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, _, err := storage.Stage(ctx, strings.NewReader("test content"))

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	if len(storage.temporary) != 0 {
		t.Errorf("expected empty temporary storage, got %d entries", len(storage.temporary))
	}
}

type cancelingReader struct {
	cancel context.CancelFunc
}

func (r cancelingReader) Read([]byte) (int, error) {
	r.cancel()
	return 0, errors.New("simulated read failure")
}

func TestMemoryStorage_Stage_ContextCanceledDuringRead(t *testing.T) {
	storage := NewMemoryStorage()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	reader := cancelingReader{cancel: cancel}

	_, _, err := storage.Stage(ctx, reader)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected %v, got %v", context.Canceled, err)
	}

	if len(storage.temporary) != 0 {
		t.Errorf("expected empty temporary storage, got %d entries", len(storage.temporary))
	}
}

func stagedContent(
	storage *MemoryStorage,
	digest artifact.Digest,
) ([]byte, bool) {
	storage.mu.RLock()
	defer storage.mu.RUnlock()

	content, ok := storage.temporary[digest]
	return bytes.Clone(content), ok
}
