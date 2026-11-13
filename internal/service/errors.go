package service

import "errors"

var (
	ErrDatabaseUnavailable             = errors.New("database unavailable")
	ErrRepositoryRequired              = errors.New("artifact repository is required")
	ErrRepositoryUpdateFailed          = errors.New("repository update failed")
	ErrBinaryStorageRequired           = errors.New("binary storage is required")
	ErrBinaryStoragePublishFailed      = errors.New("storage publish failed")
	ErrBinaryStorageDiscardStageFailed = errors.New("binary storage error failed")
	ErrArtifactNotAvailable            = errors.New("artifact is not available")
)
