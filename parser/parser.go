package parser

import (
	"bufio"
	"fmt"
	"os"
	"skill_check2/env"
	"strconv"
	"strings"
)

// ParseFile は指定されたパスのファイルを読み込み、パース結果のMapを返します。
func ParseFile(filename string) (map[string]any, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return Parse(file)
}

// Parse は io.Reader からデータを読み込み、パース結果のMapを返します。
func Parse(file *os.File) (map[string]any, error) {
	scheme, err := env.LoadSchema()
	if err != nil {
		return nil, err
	}
	result := make(map[string]any)
	scanner := bufio.NewScanner(file)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// 行頭・行末の空白を削除
		line = strings.TrimSpace(line)

		// 空行、または '#' か ';' で始まるコメント行をスキップ
		if len(line) == 0 || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// 最初に出現する '=' でキーと値に分割
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("syntax error on line %d: missing '='", lineNum)
		}

		key := strings.TrimSpace(parts[0])
		rawValue := strings.TrimSpace(parts[1])

		// スキーマに基づく検証と型変換
		typedValue, err := validateAndConvert(key, rawValue, scheme)
		if err != nil {
			return nil, fmt.Errorf("validation error on line %d: %w", lineNum, err)
		}

		// ネストしたMapへ値を挿入
		if err := insertIntoMap(result, key, typedValue); err != nil {
			return nil, fmt.Errorf("error on line %d: %w", lineNum, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// insertIntoMap はドット区切りのキーを解析し、ネストしたMapに値を設定します。
func insertIntoMap(root map[string]any, key string, value any) error {
	keys := strings.Split(key, ".")
	current := root

	// 最後のキーの手前まで、Mapを辿るか作成する
	for i := 0; i < len(keys)-1; i++ {
		k := keys[i]
		if nextNode, exists := current[k]; exists {
			if nextMap, ok := nextNode.(map[string]any); ok {
				current = nextMap
			} else {
				return fmt.Errorf("key conflict: '%s' is already set as a string, cannot be used as a dictionary", k)
			}
		} else {
			newMap := make(map[string]any)
			current[k] = newMap
			current = newMap
		}
	}

	lastKey := keys[len(keys)-1]

	// 最後のキーがすでに辞書として使われている場合はエラー
	if existingValue, exists := current[lastKey]; exists {
		if _, isMap := existingValue.(map[string]any); isMap {
			return fmt.Errorf("key conflict: '%s' is already a dictionary, cannot assign string value", lastKey)
		}
	}

	current[lastKey] = value
	return nil
}

// validateAndConvert は値がスキーマの要件を満たしているか確認し、適切な型に変換します。
func validateAndConvert(key string, value string, schema map[string]string) (interface{}, error) {
	if schema == nil {
		return value, nil // スキーマがない場合はすべて文字列として扱う
	}

	expectedType, exists := schema[key]
	if !exists {
		// スキーマに定義されていないキーをエラーにする（厳格モード）
		return nil, fmt.Errorf("key '%s' is not defined in schema", key)
	}

	switch expectedType {
	case "string":
		return value, nil
	case "bool":
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("invalid boolean value '%s' for key '%s'", value, key)
		}
		return boolVal, nil
	case "int":
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("invalid integer value '%s' for key '%s'", value, key)
		}
		return intVal, nil
	case "float":
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float value '%s' for key '%s'", value, key)
		}
		return floatVal, nil
	default:
		return nil, fmt.Errorf("unsupported type '%s' in schema for key '%s'", expectedType, key)
	}
}
