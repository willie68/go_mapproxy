package assets

import (
	"bytes"
	_ "embed"
	"io"
)

//go:embed empty.png
var emptyPNG []byte

func EmptyPNG() io.ReadCloser {
	return io.NopCloser(io.Reader(bytes.NewReader(emptyPNG)))
}
