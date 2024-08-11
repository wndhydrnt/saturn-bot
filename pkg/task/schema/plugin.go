package schema

import "path/filepath"

func (p *Plugin) PathAbs(taskPath string) string {
	if filepath.IsAbs(p.Path) {
		return p.Path
	}

	dir := filepath.Dir(taskPath)
	return filepath.Join(dir, p.Path)
}
