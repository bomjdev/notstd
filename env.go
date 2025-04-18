package notstd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func LoadDotEnv() error {
	return LoadEnvFromFile(".env")
}

func LoadEnvFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first '='
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // or return error if strict
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove optional surrounding quotes
		value = strings.Trim(value, `"'`)

		if err = os.Setenv(key, value); err != nil {
			return fmt.Errorf("set env: %w", err)
		}
	}

	return scanner.Err()
}
