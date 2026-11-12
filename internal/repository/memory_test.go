package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/istvzsig/artifix/internal/artifact"
)

func TestMemoryRepository_Create_Success(t *testing.T) {
	repository := NewMemoryRepository()

	coords, err := artifact.NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	a, err := artifact.NewArtifact(coords, artifact.Digest{}, 1024, "application/zip")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := repository.Create(context.Background(), a); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestMemoryRepository_Create_DuplicateCoordinates(t *testing.T) {
	repository := NewMemoryRepository()

	coords, err := artifact.NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	a1, err := artifact.NewArtifact(coords, artifact.Digest{}, 1024, "application/zip")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := repository.Create(context.Background(), a1); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var differentDigest artifact.Digest
	differentDigest[0] = 1

	a2, err := artifact.NewArtifact(coords, differentDigest, 1024, "application/zip")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = repository.Create(context.Background(), a2)
	if !errors.Is(err, ErrArtifactAlreadyExists) {
		t.Fatalf("expected ErrArtifactAlreadyExists, got %v", err)
	}
}

func TestMemoryRepository_Create_CanceledContext(t *testing.T) {
	repository := NewMemoryRepository()

	coords, err := artifact.NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	a, err := artifact.NewArtifact(coords, artifact.Digest{}, 1024, "application/zip")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err = repository.Create(ctx, a)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestMemoryRepository_Get_Success(t *testing.T) {
	repository := NewMemoryRepository()

	coords, err := artifact.NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	a, err := artifact.NewArtifact(coords, artifact.Digest{}, 1024, "application/zip")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := repository.Create(context.Background(), a); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got, err := repository.Get(context.Background(), coords)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if a.ID() != got.ID() {
		t.Fatalf("expected artifact IDs to match, got %q and %q", a.ID(), got.ID())
	}

	if a.Digest() != got.Digest() {
		t.Fatalf("expected artifact digests to match, got %v and %v", a.Digest(), got.Digest())
	}

	if a.Status() != got.Status() {
		t.Fatalf("expected artifact statuses to match, got %q and %q", a.Status(), got.Status())
	}

	if a.Size() != got.Size() {
		t.Fatalf("expected artifact sizes to match, got %d and %d", a.Size(), got.Size())
	}
}

func TestMemoryRepository_Get_NotFound(t *testing.T) {
	repository := NewMemoryRepository()

	coords, err := artifact.NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	_, err = repository.Get(context.Background(), coords)
	if !errors.Is(err, ErrArtifactNotFound) {
		t.Fatalf("expected ErrArtifactNotFound, got %v", err)
	}
}

func TestMemoryRepository_Get_CanceledContext(t *testing.T) {
	repository := NewMemoryRepository()

	coords, err := artifact.NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err = repository.Get(ctx, coords)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestMemoryRepository_Update_Success(t *testing.T) {
	repository := NewMemoryRepository()

	coords, err := artifact.NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	a, err := artifact.NewArtifact(coords, artifact.Digest{}, 1024, "application/zip")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := repository.Create(context.Background(), a); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := a.MarkAvailable(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if errors.Is(err, artifact.ErrInvalidStatusTransition) {
		t.Fatalf("expected no artifact.ErrInvalidStatusTransition error")
	}

	if err := repository.Update(context.Background(), a); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got, err := repository.Get(context.Background(), coords)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.Status() != artifact.StatusAvailable {
		t.Fatalf("expected status %q, got %q",
			artifact.StatusAvailable,
			got.Status(),
		)
	}
}

func TestMemoryRepository_Update_NotFound(t *testing.T) {
	repository := NewMemoryRepository()

	coords, err := artifact.NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	a, err := artifact.NewArtifact(coords, artifact.Digest{}, 1024, "application/zip")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = repository.Update(context.Background(), a)
	if !errors.Is(err, ErrArtifactNotFound) {
		t.Fatalf("expected ErrArtifactNotFound, got %v", err)
	}
}

func TestMemoryRepository_Update_IDMismatch(t *testing.T) {
	repository := NewMemoryRepository()

	coords, err := artifact.NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	a1, err := artifact.NewArtifact(coords, artifact.Digest{}, 1024, "application/zip")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := repository.Create(context.Background(), a1); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	a2, err := artifact.NewArtifact(coords, artifact.Digest{}, 1024, "application/zip")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = repository.Update(context.Background(), a2)
	if !errors.Is(err, ErrArtifactIDMismatch) {
		t.Fatalf("expected ErrArtifactIDMismatch, got %v", err)
	}

	got, err := repository.Get(context.Background(), coords)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.ID() != a1.ID() {
		t.Fatalf("expected original ID %q, got %q", a1.ID(), got.ID())
	}
}

func TestMemoryRepository_Update_CanceledContext(t *testing.T) {
	repository := NewMemoryRepository()

	coords, err := artifact.NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	a, err := artifact.NewArtifact(coords, artifact.Digest{}, 1024, "application/zip")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := repository.Create(context.Background(), a); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err = repository.Update(ctx, a)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	got, err := repository.Get(context.Background(), coords)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.ID() != a.ID() {
		t.Fatalf("expected ID %q, got %q", a.ID(), got.ID())
	}
}

func TestMemoryRepository_Update_PreservesImmutableFields(t *testing.T) {
	repository := NewMemoryRepository()

	coords, err := artifact.NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	digest := artifact.Digest{}
	digest[0] = 42

	a, err := artifact.NewArtifact(coords, digest, 1024, "application/zip")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	originalID := a.ID()
	originalDigest := a.Digest()
	originalSize := a.Size()
	originalMediaType := a.MediaType()
	originalCoords := a.Coordinates()

	if err := repository.Create(context.Background(), a); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := a.MarkAvailable(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := repository.Update(context.Background(), a); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got, err := repository.Get(context.Background(), coords)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.ID() != originalID {
		t.Errorf("expected ID %q, got %q", originalID, got.ID())
	}

	if got.Digest() != originalDigest {
		t.Errorf("expected digest %v, got %v", originalDigest, got.Digest())
	}

	if got.Size() != originalSize {
		t.Errorf("expected size %d, got %d", originalSize, got.Size())
	}

	if got.MediaType() != originalMediaType {
		t.Errorf("expected media type %q, got %q", originalMediaType, got.MediaType())
	}

	if got.Coordinates() != originalCoords {
		t.Errorf("expected coordinates %v, got %v", originalCoords, got.Coordinates())
	}

	if got.Status() != artifact.StatusAvailable {
		t.Errorf("expected status %q, got %q", artifact.StatusAvailable, got.Status())
	}
}

func TestMemoryRepository_List_Success(t *testing.T) {
	repository := NewMemoryRepository()

	// Artifact 1
	coords, err := artifact.NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	a, err := artifact.NewArtifact(
		coords,
		artifact.Digest{},
		100,
		"application/octet-stream",
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := repository.Create(context.Background(), a); err != nil {
		t.Fatal(err)
	}

	// Artifact 2
	coords2, err := artifact.NewCoordinates("acme", "api", "2.0.0")
	if err != nil {
		t.Fatal(err)
	}

	a2, err := artifact.NewArtifact(
		coords2,
		artifact.Digest{},
		512,
		"application/zip",
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := repository.Create(context.Background(), a2); err != nil {
		t.Fatal(err)
	}

	// Artifact 3
	coords3, err := artifact.NewCoordinates("acme", "worker", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	a3, err := artifact.NewArtifact(
		coords3,
		artifact.Digest{},
		1024,
		"application/octet-stream",
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := repository.Create(context.Background(), a3); err != nil {
		t.Fatal(err)
	}

	result, err := repository.List(context.Background(), "acme", "api")
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 artifacts, got %d", len(result))
	}

	for _, a := range result {
		coords := a.Coordinates()

		if coords.Namespace() != "acme" {
			t.Errorf("expected namespace acme, got %s", coords.Namespace())
		}

		if coords.Name() != "api" {
			t.Errorf("expected name api, got %s", coords.Name())
		}
	}
}

func TestMemoryRepository_List_Empty(t *testing.T) {
	repository := NewMemoryRepository()

	result, err := repository.List(context.Background(), "acme", "api")
	if err != nil {
		t.Fatal(err)
	}

	if result == nil {
		t.Fatal("expected non-nil empty slice")
	}

	if len(result) != 0 {
		t.Fatalf("expected empty artifacts, got %d", len(result))
	}
}

func TestMemoryRepository_List_DifferentPackage(t *testing.T) {
	repository := NewMemoryRepository()

	coords, err := artifact.NewCoordinates("acme", "worker", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	a, err := artifact.NewArtifact(
		coords,
		artifact.Digest{},
		100,
		"application/octet-stream",
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := repository.Create(context.Background(), a); err != nil {
		t.Fatal(err)
	}

	result, err := repository.List(context.Background(), "acme", "api")
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 0 {
		t.Fatalf("expected zero result, got %d", len(result))
	}
}

func TestMemoryRepository_List_CanceledContext(t *testing.T) {
	repository := NewMemoryRepository()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	result, err := repository.List(ctx, "acme", "api")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}
}
