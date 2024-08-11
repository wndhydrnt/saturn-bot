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
func Read(path string) ([]Task, []hash.Hash, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("read task file '%s': %w", path, err)
	}
	dec := yaml.NewDecoder(bytes.NewReader(b))
	var result []Task
	var hashes []hash.Hash
	// Decode in a loop to account for multiple documents in one file
	for {
		var task Task
		err := dec.Decode(&task)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, nil, fmt.Errorf("decode task file from YAML: %w", err)
		}

		// Encode to YAML again. Calculates the hash of the task,
		// not of the file, which can contain more than one task.
		checksum := sha256.New()
		enc := yaml.NewEncoder(checksum)
		if err := enc.Encode(&task); err != nil {
			return nil, nil, fmt.Errorf("encode task for hash: %w", err)
		}

		if err := enc.Close(); err != nil {
			return nil, nil, fmt.Errorf("close yaml encoder for hash: %w", err)
		}

		if err := Validate(&task); err != nil {
			return nil, nil, fmt.Errorf("validation failed: %w", err)
		}

		for _, plugin := range task.Plugins {
			pluginFile, err := os.Open(plugin.PathAbs(path))
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
		hashes = append(hashes, checksum)
	}

	return result, hashes, nil
}
