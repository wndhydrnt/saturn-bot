package plugin

import (
	"bytes"
)

type stdioAdapter struct {
	name      string
	onMessage StdioHandler
}

// Write implements io.Writer
// It trims space characters at the begging and end of the received data.
func (s *stdioAdapter) Write(d []byte) (int, error) {
	if s.onMessage != nil && len(bytes.TrimSpace(d)) != 0 {
		s.onMessage(s.name, bytes.TrimSpace(d))
	}

	return len(d), nil
}
