package schema

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// Read return all Tasks from the file at `path`.
// It also calculates the hash of the file.
func Read(path string) ([]Task, hash.Hash, error) {
	var result []Task
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("read task file '%s': %w", path, err)
	}

	checksum := sha256.New()
	_, _ = checksum.Write(b)
	dec := yaml.NewDecoder(bytes.NewReader(b))
	// Decode in a loop to account for multiple documents in one file
	for {
		var task Task
		err := dec.Decode(&task)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, nil, fmt.Errorf("decode task file '%s' from YAML: %w", path, err)
		}

		err = Validate(&task)
		if err != nil {
			return nil, nil, fmt.Errorf("validation failed: %w", err)
		}

		for _, plugin := range task.Plugins {
			pluginFile, err := os.Open(plugin.Path)
			if err != nil {
				return nil, nil, fmt.Errorf("open plugin file: %w", err)
			}

			defer pluginFile.Close()
			_, err = io.Copy(checksum, pluginFile)
			if err != nil {
				return nil, nil, fmt.Errorf("calculating hash of plugin: %w", err)
			}
		}

		result = append(result, task)
	}

	return result, checksum, nil
}
