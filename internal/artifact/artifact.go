package artifact

import (
	"crypto/sha256"
	"time"
	"uuid"
)

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusAvailable Status = "AVAILABLE"
	StatusDeleted   Status = "DELETED"
)

type Digest [sha256.Size]byte

type Artifact struct {
	ID         uuid.UUID
	Namespace  string
	Name       string
	Version    string
	Digest     Digest
	Size       int64
	MediaType  string
	CreatedAt  time.Time
	DeletedAt  time.Time
	StorageKey string
}
