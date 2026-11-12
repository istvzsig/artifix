package repository

import "errors"

var (
	ErrArtifactAlreadyExists = errors.New("artifact already exists")
	ErrArtifactNotFound      = errors.New("artifact not found")
	ErrArtifactIDMismatch    = errors.New("artifact ID mismatch")
)
