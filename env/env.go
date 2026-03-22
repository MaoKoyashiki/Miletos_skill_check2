package env

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const schemeDef = "./env/scheme.conf"

// LoadSchema はスキーマ定義を読み込み、キーと型のマッピングを返します。
func LoadSchema() (map[string]string, error) {
	path, err := filepath.Abs(schemeDef)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	schema := make(map[string]string)
	scanner := bufio.NewScanner(f)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if len(line) == 0 || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("schema syntax error on line %d: missing '='", lineNum)
		}
		schema[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return schema, scanner.Err()
}
