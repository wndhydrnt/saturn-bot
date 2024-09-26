package plugin

import (
	"bytes"
)

type stdioAdapter struct {
	name      string
	onMessage func(pluginName string, msg string)
}

// Write implements io.Writer
func (s *stdioAdapter) Write(d []byte) (int, error) {
	if s.onMessage != nil && len(bytes.TrimSpace(d)) != 0 {
		s.onMessage(s.name, string(d))
	}

	return len(d), nil
}
