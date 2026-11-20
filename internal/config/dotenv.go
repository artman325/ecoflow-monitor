package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

var ErrInvalidLine = errors.New("dotenv: incorrect line format")

var ErrEnvFileNotFound = errors.New("dotenv: .env file not found")
var ErrEnvFileUnreadable = errors.New("dotenv: cannot read .env file")

const bom = "\uFEFF"

func parse(path string) (map[string]string, error) {
	m := make(map[string]string)

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("%w: %w", ErrEnvFileNotFound, err)
		}
		return nil, fmt.Errorf("%w: %w", ErrEnvFileUnreadable, err)
	}
	content := strings.ReplaceAll(string(data), "\r\n", "\n")

	content = strings.TrimPrefix(content, bom)

	lines := strings.Split(content, "\n")

	for i, line := range lines {

		line = strings.TrimSpace(line)
		// пропускаем пустые строки и строки которые закомментированы
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			return nil, fmt.Errorf("%w: line %d (%q)", ErrInvalidLine, i+1, line)
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		m[key] = value
	}

	return m, nil
}
