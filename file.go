package tailpipe

import (
	"io"
	"sync"
)

// File represents an open normal file. A File is effectively of infinite
// length; all reads to the file will block until data are available,
// even if EOF on the underlying file is reached.
//
// The tailpipe package will attempt to detect when a file has been
// rotated. Programs that wish to be notified when such a rotation occurs
// should receive from the Rotated channel.
type File struct {
	r       io.Reader
	Rotated <-chan struct{}
	mu      sync.RWMutex
	rc      chan struct{}
}
