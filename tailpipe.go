// Package tailpipe allows for reading normal files indefinitely.  With the
// tailpipe package, code that uses the standard library's io.Reader
// interface can be transparently adapted to receive future updates to
// normal files.
package tailpipe

import (
	"io"
	"os"
	"time"
)

// Read reads up to len(p) bytes into p. If end-of-file is reached, Read
// will block until new data are available. Read returns the number of
// bytes read and any errors other than io.EOF.
//
// If the underlying stream is an *os.File, Read will attempt to detect
// if it has been replaced, such as during log rotation. If so, Read will
// re-open the file at the original path provided to Open. Re-opening
// does not occur until the old file is exhausted.
func (f *File) Read(p []byte) (n int, err error) {
	tick := time.Tick(time.Millisecond * 100)

	for {
		n, err = f.r.Read(p)
		if n == 0 && err == io.EOF {
			<-tick
			if file, ok := f.r.(*os.File); ok {
				if file, ok = newFile(file); ok {
					select {
					case f.rc <- struct{}{}:
					default:
					}
					f.r = file
				}
			} else {
				return 0, ErrNotSupported
			}
		} else {
			break
		}
	}
	if err == io.EOF {
		return n, nil
	}
	return n, err
}

func newFile(oldfile *os.File) (*os.File, bool) {
	// NOTE(dwisiswant0): It is recommended to use os.Stat on
	// the file descriptor of the open file to check for rotation,
	// this eliminates the race condition that currently exists
	// in the code and makes the rotation detection more reliable.
	newfile, err := os.Open(oldfile.Name())
	if err != nil {
		// NOTE(dwisiswant0): if os.Open returns an error, it could
		// be because the file is no longer present or it could be
		// in between rotations. A timeout can be added for more
		// complex handling of these cases.
		return nil, false
	}

	oldstat, err := oldfile.Stat()
	if err != nil {
		oldfile.Close()
		return newfile, true
	}

	newstat, err := newfile.Stat()
	if err != nil {
		newfile.Close()
		return oldfile, false
	}

	if oldstat.Size() != newstat.Size() || oldstat.ModTime() != newstat.ModTime() {
		oldfile.Close()
		return newfile, true
	}
	newfile.Close()

	return oldfile, false
}

// Open opens the given file for reading.
func Open(path string) (*File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if !fi.Mode().IsRegular() {
		return nil, ErrNotRegularFile
	}

	return Follow(f), err
}

// Follow converts an existing io.Reader to a tailpipe.File. This can
// be useful when opening files with special permissions. In general,
// the behavior of a tailpipe.File is only suitable for normal files;
// using Follow on an io.Pipe, net.Conn, or other non-file stream will
// yield undesirable results.
func Follow(r io.Reader) *File {
	// BUG(droyo): a File's Rotated channel should not be relied upon to
	// provide an accurate count of file rotations; because the tailpipe
	// will only perform a non-blocking send on the Rotated channel,
	// a goroutine may miss a new notification while it is responding
	// to a previous notification. This is addressed by buffering the
	// channel, but can still be a problem with files that are rotated
	// very frequently.
	f := &File{r: r, rc: make(chan struct{}, 1)}
	f.Rotated = f.rc
	return f
}

// Name returns the name of the underlying file, if available. If the
// underlying stream does not have a name, Name returns an empty string.
func (f *File) Name() string {
	type named interface {
		Name() string
	}
	if v, ok := f.r.(named); ok {
		return v.Name()
	}
	return ""
}

// Seek calls Seek on the underlying stream. If the underlying stream does
// not provide a Seek method, ErrNotSupported is returned.
func (f *File) Seek(offset int64, whence int) (int64, error) {
	if v, ok := f.r.(io.Seeker); ok {
		return v.Seek(offset, whence)
	}
	return 0, ErrNotSupported
}

// Close closes the underlying file or stream. It also has the
// side effect of closing the File's Rotated channel. Close
// is safe to call multiple times.
func (f *File) Close() error {
	f.mu.Lock()
	if f.rc != nil {
		close(f.rc)
		f.rc = nil
	}

	f.mu.Unlock()
	if v, ok := f.r.(io.ReadCloser); ok {
		return v.Close()
	}

	return ErrNotSupported
}
