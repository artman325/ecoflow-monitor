package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

var ErrInvalidLine = errors.New("dotenv: incorrect line format")

const bom = "\uFEFF"

func parse(path string) (map[string]string, error) {
	m := make(map[string]string)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("dotenv: %w: error open file", err)
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
