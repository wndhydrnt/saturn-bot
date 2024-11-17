package schema

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"

	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"gopkg.in/yaml.v3"
)

type ReadResult struct {
	Hash   hash.Hash
	Path   string
	Sha256 string
	Task   Task
}

// Read return all Tasks from the file at `path`.
// It also calculates the hash of the file.
func Read(path string) ([]ReadResult, error) {
	var results []ReadResult
	globs, err := filepath.Glob(path)
	if err != nil {
		return nil, fmt.Errorf("globbing task file '%s': %w", path, err)
	}

	for _, name := range globs {
		readResults, err := readFile(name)
		if err != nil {
			return nil, err
		}

		results = append(results, readResults...)
	}

	return results, nil
}

func readFile(path string) ([]ReadResult, error) {
	log.Log().Debugf("Reading task file %s", path)
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read task file '%s': %w", path, err)
	}
	var result []ReadResult
	dec := yaml.NewDecoder(bytes.NewReader(b))
	// Decode in a loop to account for multiple documents in one file
	for {
		var task Task
		err := dec.Decode(&task)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("decode task from YAML file %s: %w", path, err)
		}

		// Encode to YAML again. Calculates the hash of the task,
		// not of the file, which can contain more than one task.
		checksum := sha256.New()
		enc := yaml.NewEncoder(checksum)
		if err := enc.Encode(&task); err != nil {
			return nil, fmt.Errorf("encode task for hash: %w", err)
		}

		if err := enc.Close(); err != nil {
			return nil, fmt.Errorf("close yaml encoder for hash: %w", err)
		}

		if err := Validate(&task); err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}

		for _, plugin := range task.Plugins {
			pluginFile, err := os.Open(plugin.PathAbs(path))
			if err != nil {
				return nil, fmt.Errorf("open plugin file: %w", err)
			}

			defer pluginFile.Close()
			_, err = io.Copy(checksum, pluginFile)
			if err != nil {
				return nil, fmt.Errorf("calculating hash of plugin: %w", err)
			}
		}

		log.Log().Debugf("Found task %s", task.Name)
		result = append(result, ReadResult{
			Hash:   checksum,
			Path:   path,
			Sha256: hex.EncodeToString(checksum.Sum(nil)),
			Task:   task,
		})
	}

	return result, nil
}
