package storage

import "errors"

var (
	ErrFailedToReadContent   = errors.New("failed to read content")
	ErrArtifactTooLarge      = errors.New("artifact exceeds maximum allowed size")
	ErrStagedContentNotFound = errors.New("staged content not found")
)
