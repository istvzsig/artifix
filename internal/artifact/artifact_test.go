package artifact

import (
	"errors"
	"testing"

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
