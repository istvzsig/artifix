package artifact

import "errors"

var (
	ErrCoordinatesCannotBeZero = errors.New("artifact coordinates cannot be zero")
	ErrSizeCannotBeNegative    = errors.New("artifact size cannot be negative")
	ErrInvalidMediaType        = errors.New("invalid artifact media type")
	ErrInvalidStatusTransition = errors.New("invalid artifact status transition")
)
