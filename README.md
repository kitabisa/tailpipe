[![GoDoc](https://godoc.org/github.com/kitabisa/tailpipe?status.svg)](https://godoc.org/github.com/kitabisa/tailpipe)

> This version of https://github.com/droyo/tailpipe has been modified to include additional improvements.

The tailpipe package provides an implementation of `tail -F` for Go. In
effect, it produces files that are infinite length, and re-opens the
file when it is changed, such as during log rotation. It can be used
for programs that need to watch log files indefinitely.

There are several `tail` packages for Go available. However, most
packages do not preserve the io.Reader interface, instead opting for
a line-oriented approach involving `chan string`. By preserving the
`io.Reader` interface, this package composes better with the standard
library.
