package tailpipe

import "errors"

var (
	// ErrNotSupported is returned when an operation is attempted on an underlying stream that does not support it.
	ErrNotSupported = errors.New("operation not supported by underlying stream")

	// ErrNotRegularFile is returned when a provided file is not a regular file.
	ErrNotRegularFile = errors.New("provided file is not a regular file")
)
