package artifact

import (
	"crypto/sha256"
	"fmt"
	"mime"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusAvailable Status = "AVAILABLE"
	StatusDeleted   Status = "DELETED"
)

type Digest [sha256.Size]byte

type Artifact struct {
	id        uuid.UUID
	coords    Coordinates
	digest    Digest
	size      int64
	mediaType string
	status    Status
	createdAt time.Time
	deletedAt time.Time
}

func NewArtifact(
	coords Coordinates,
	digest Digest,
	size int64,
	mediaType string,
) (Artifact, error) {
	if !coords.IsValid() {
		return Artifact{}, ErrCoordinatesCannotBeZero
	}

	if size < 0 {
		return Artifact{}, ErrSizeCannotBeNegative
	}

	parsedType, _, err := mime.ParseMediaType(mediaType)
	if err != nil || !strings.Contains(parsedType, "/") {
		return Artifact{}, fmt.Errorf("%w: %q", ErrInvalidMediaType, mediaType)
	}

	artifact := Artifact{
		id:        uuid.New(),
		coords:    coords,
		digest:    digest,
		size:      size,
		mediaType: parsedType,
		status:    StatusPending,
		createdAt: time.Now().UTC(),
	}

	return artifact, nil
}

func (a *Artifact) MarkAvailable() error {
	if a.status != StatusPending {
		return fmt.Errorf(
			"%w: cannot mark artifact available with status %s",
			ErrInvalidStatusTransition,
			a.status,
		)
	}

	a.status = StatusAvailable

	return nil
}

func (a *Artifact) Delete() error {
	if a.status != StatusAvailable {
		return fmt.Errorf(
			"%w: cannot delete artifact with status %s",
			ErrInvalidStatusTransition,
			a.status,
		)
	}

	a.status = StatusDeleted
	a.deletedAt = time.Now().UTC()

	return nil
}
