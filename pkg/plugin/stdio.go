package plugin

import "fmt"

const (
	streamStderr = "stderr"
	streamStdout = "stdout"
)

type stdioAdapter struct {
	name      string
	stream    string
	onMessage func(string)
}

// Write implements io.Writer
func (s *stdioAdapter) Write(d []byte) (int, error) {
	if s.onMessage != nil && string(d) != "\n" {
		s.onMessage(fmt.Sprintf("[PLUGIN %s %s] %s", s.name, s.stream, string(d)))
	}

	return len(d), nil
}
