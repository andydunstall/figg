package utils

import (
	"io"
)

type NetworkConnection interface {
	io.Reader
	io.Writer
	io.Closer
}
