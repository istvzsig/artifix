package artifact

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewArtifact(t *testing.T) {
	coords, err := NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	got, err := NewArtifact(coords, Digest{}, 1024, "application/zip")
	if err != nil {
		t.Fatalf("NewArtifact() error = %v", err)
	}

	if got.id == uuid.Nil {
		t.Error("expected artifact ID to be generated")
	}

	if got.size != 1024 {
		t.Errorf("expected size 1024, got %d", got.size)
	}

	if got.status != StatusPending {
		t.Errorf("expected initial status PENDING, got %s", got.status)
	}

	if !got.deletedAt.IsZero() {
		t.Errorf("expected zero deletion time, got %v", got.deletedAt)
	}
}

func TestNewArtifact_InvalidSize(t *testing.T) {
	coords, err := NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewArtifact(coords, Digest{}, -1024, "application/zip")

	if !errors.Is(err, ErrSizeCannotBeNegative) {
		t.Fatalf("expected ErrSizeCannotBeNegative, got %v", err)
	}
}

func TestNewArtifact_InvalidMediaType(t *testing.T) {
	coords, err := NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewArtifact(coords, Digest{}, 1024, "not-a-media-type")
	if !errors.Is(err, ErrInvalidMediaType) {
		t.Fatalf("expected ErrInvalidMediaType, got %v", err)
	}
}

func TestNewArtifact_NormalizesMediaType(t *testing.T) {
	coords, err := NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	got, err := NewArtifact(coords, Digest{}, 1024, "Application/Zip")
	if err != nil {
		t.Fatalf("NewArtifact() error = %v", err)
	}

	want := "application/zip"
	if got.mediaType != want {
		t.Errorf("expected media type %q, got %q", want, got.mediaType)
	}
}

func TestNewArtifact_ZeroCoordinates(t *testing.T) {
	var coords Coordinates

	_, err := NewArtifact(coords, Digest{}, 1024, "Application/Zip")
	if !errors.Is(err, ErrCoordinatesCannotBeZero) {
		t.Fatalf("expected ErrCoordinatesCannotBeZero, got %v", err)
	}
}

func TestNewArtifact_ValidationOrder(t *testing.T) {
	var coords Coordinates

	_, err := NewArtifact(coords, Digest{}, -1024, "Application/Zip")
	if !errors.Is(err, ErrCoordinatesCannotBeZero) {
		t.Fatalf("expected ErrCoordinatesCannotBeZero, got %v", err)
	}
}

func TestArtifact_MarkAvailable(t *testing.T) {
	coords, err := NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	got, err := NewArtifact(coords, Digest{}, 1024, "Application/Zip")
	if err != nil {
		t.Fatalf("NewArtifact() error = %v", err)
	}

	err = got.MarkAvailable()
	if err != nil {
		t.Fatalf("MarkAvailable() error = %v", err)
	}

	if got.status != StatusAvailable {
		t.Fatalf("expected StatusAvailable, got %v", got.status)
	}
}

func TestArtifact_MarkAvailable_InvalidTransition(t *testing.T) {
	coords, err := NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	got, err := NewArtifact(coords, Digest{}, 1024, "Application/Zip")
	if err != nil {
		t.Fatalf("NewArtifact() error = %v", err)
	}

	err = got.MarkAvailable()
	if err != nil {
		t.Fatalf("first MarkAvailable() error = %v", err)
	}

	err = got.MarkAvailable()
	if !errors.Is(err, ErrInvalidStatusTransition) {
		t.Fatalf("second MarkAvailable() error = %v", err)
	}

	if got.status != StatusAvailable {
		t.Fatalf("expected status to remain AVAILABLE, got %v", got.status)
	}
}

func TestArtifact_Delete(t *testing.T) {
	coords, err := NewCoordinates("acme", "api", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	got, err := NewArtifact(coords, Digest{}, 1024, "Application/Zip")
	if err != nil {
		t.Fatalf("NewArtifact() error = %v", err)
	}

	if err := got.MarkAvailable(); err != nil {
		t.Fatalf("MarkAvailable() error = %v", err)
	}

	if err := got.Delete(); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if got.status != StatusDeleted {
		t.Fatalf("expected StatusDeleted, got %v", got.status)
	}

	if got.deletedAt.IsZero() {
		t.Fatal("expected deletedAt to be set")
	}

	if got.deletedAt.Location() != time.UTC {
		t.Fatalf("expected UTC deletion timestamp, got %v", got.deletedAt.Location())
	}
}
