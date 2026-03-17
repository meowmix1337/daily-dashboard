package service

import (
	"io"
)

// maxBodySize caps upstream HTTP response bodies at 1 MB to prevent OOM from
// malicious or misbehaving external APIs.
const maxBodySize = 1 << 20 // 1 MB

// readBody reads at most maxBodySize bytes from r.
// It prevents unbounded memory usage when consuming external API responses.
func readBody(r io.Reader) ([]byte, error) {
	return io.ReadAll(io.LimitReader(r, maxBodySize))
}
