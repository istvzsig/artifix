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

	if err := repository.Create(t.Context(), a); err != nil {
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

	if err := repository.Create(t.Context(), a1); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var differentDigest artifact.Digest
	differentDigest[0] = 1

	a2, err := artifact.NewArtifact(coords, differentDigest, 1024, "application/zip")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = repository.Create(t.Context(), a2)
	if !errors.Is(err, ErrArtifactAlreadyExists) {
		t.Fatalf("expected %v, got %v", ErrArtifactAlreadyExists, err)
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
		t.Fatalf("expected %v, got %v", context.Canceled, err)
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

	if err := repository.Create(t.Context(), a); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got, err := repository.Get(t.Context(), coords)
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

	_, err = repository.Get(t.Context(), coords)
	if !errors.Is(err, ErrArtifactNotFound) {
		t.Fatalf("expected %v, got %v", ErrArtifactNotFound, err)
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

	if err := repository.Create(t.Context(), a); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := a.MarkAvailable(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if errors.Is(err, artifact.ErrInvalidStatusTransition) {
		t.Fatalf("expected no %v error", artifact.ErrInvalidStatusTransition)
	}

	if err := repository.Update(t.Context(), a); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got, err := repository.Get(t.Context(), coords)
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

	err = repository.Update(t.Context(), a)
	if !errors.Is(err, ErrArtifactNotFound) {
		t.Fatalf("expected %v, got %v", ErrArtifactNotFound, err)
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

	if err := repository.Create(t.Context(), a1); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	a2, err := artifact.NewArtifact(coords, artifact.Digest{}, 1024, "application/zip")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = repository.Update(t.Context(), a2)
	if !errors.Is(err, ErrArtifactIDMismatch) {
		t.Fatalf("expected %v, got %v", ErrArtifactIDMismatch, err)
	}

	got, err := repository.Get(t.Context(), coords)
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

	if err := repository.Create(t.Context(), a); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err = repository.Update(ctx, a)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected %v, got %v", context.Canceled, err)
	}

	got, err := repository.Get(t.Context(), coords)
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

	if err := repository.Create(t.Context(), a); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := a.MarkAvailable(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := repository.Update(t.Context(), a); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got, err := repository.Get(t.Context(), coords)
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

	if err := repository.Create(t.Context(), a); err != nil {
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

	if err := repository.Create(t.Context(), a2); err != nil {
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

	if err := repository.Create(t.Context(), a3); err != nil {
		t.Fatal(err)
	}

	result, err := repository.List(t.Context(), "acme", "api")
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

	result, err := repository.List(t.Context(), "acme", "api")
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

	if err := repository.Create(t.Context(), a); err != nil {
		t.Fatal(err)
	}

	result, err := repository.List(t.Context(), "acme", "api")
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
		t.Fatalf("expected %v, got %v", context.Canceled, err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}
}

func TestMemoryRepository_Delete_Success(t *testing.T) {
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

	if err := repository.Create(t.Context(), a); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := a.MarkAvailable(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := repository.Update(t.Context(), a); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := repository.Delete(t.Context(), coords); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got, err := repository.Get(t.Context(), coords)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.Status() != artifact.StatusDeleted {
		t.Fatalf("expected status %q, got %q",
			artifact.StatusDeleted,
			got.Status(),
		)
	}

	if got.DeletedAt().IsZero() {
		t.Fatal("expected non-zero deletion timestamp")
	}
}

func TestMemoryRepository_Delete_NotFound(t *testing.T) {
	repository := NewMemoryRepository()

	coords, err := artifact.NewCoordinates("acme", "worker", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	err = repository.Delete(t.Context(), coords)
	if !errors.Is(err, ErrArtifactNotFound) {
		t.Fatalf("expected error %v, got %v", ErrArtifactNotFound, err)
	}
}

func TestMemoryRepository_Delete_CanceledContext(t *testing.T) {
	repository := NewMemoryRepository()

	coords, err := artifact.NewCoordinates("acme", "worker", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err = repository.Delete(ctx, coords)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected error %v, got %v", context.Canceled, err)
	}
}

func TestMemoryRepository_Delete_AlreadyDeleted(t *testing.T) {
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

	if err := repository.Create(t.Context(), a); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := a.MarkAvailable(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := repository.Update(t.Context(), a); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := repository.Delete(t.Context(), coords); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got, err := repository.Get(t.Context(), coords)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.Status() != artifact.StatusDeleted {
		t.Fatalf("expected status %q, got %q",
			artifact.StatusDeleted,
			got.Status(),
		)
	}

	if got.DeletedAt().IsZero() {
		t.Fatal("expected non-zero deletion timestamp")
	}

	err = repository.Delete(t.Context(), coords)
	if !errors.Is(err, artifact.ErrInvalidStatusTransition) {
		t.Fatalf(
			"expected error %v, got %v",
			artifact.ErrInvalidStatusTransition,
			err,
		)
	}
}

func TestMemoryRepository_Delete_PreservesArtifactMetadata(t *testing.T) {
	repository := NewMemoryRepository()

	coords, err := artifact.NewCoordinates("acme", "worker", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	digest := artifact.Digest{1, 2, 3, 4}

	a, err := artifact.NewArtifact(
		coords,
		digest,
		1024,
		"application/octet-stream",
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := a.MarkAvailable(); err != nil {
		t.Fatal(err)
	}

	if err := repository.Create(t.Context(), a); err != nil {
		t.Fatal(err)
	}

	if err := repository.Delete(t.Context(), coords); err != nil {
		t.Fatal(err)
	}

	got, err := repository.Get(t.Context(), coords)
	if err != nil {
		t.Fatal(err)
	}

	if got.ID() != a.ID() {
		t.Errorf("expected ID %v, got %v", a.ID(), got.ID())
	}

	if got.Coordinates() != a.Coordinates() {
		t.Errorf("expected coordinates %v, got %v", a.Coordinates(), got.Coordinates())
	}

	if got.Digest() != a.Digest() {
		t.Errorf("expected digest %v, got %v", a.Digest(), got.Digest())
	}

	if got.Size() != a.Size() {
		t.Errorf("expected size %d, got %d", a.Size(), got.Size())
	}

	if got.MediaType() != a.MediaType() {
		t.Errorf("expected media type %q, got %q", a.MediaType(), got.MediaType())
	}

	if !got.CreatedAt().Equal(a.CreatedAt()) {
		t.Errorf("expected creation time %v, got %v", a.CreatedAt(), got.CreatedAt())
	}

	if got.Status() != artifact.StatusDeleted {
		t.Errorf("expected status %q, got %q", artifact.StatusDeleted, got.Status())
	}

	if got.DeletedAt().IsZero() {
		t.Error("expected non-zero deletion timestamp")
	}
}
