package config

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/gilankpam/openipc-gs-web/internal/models"
)

// LoadAlink loads Alink configuration from a key-value file
func (s *ServiceConfig) LoadAlink() (*models.AlinkConfig, error) {
	file, err := os.Open(s.AlinkPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open alink config: %w", err)
	}
	defer file.Close()

	config := &models.AlinkConfig{}

	// Map tags to field map for quick lookup
	// We need to set values on 'config'
	v := reflect.ValueOf(config).Elem()
	t := v.Type()

	tagToField := make(map[string]int)
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("conf")
		if tag != "" {
			tagToField[tag] = i
		}
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		if fieldIdx, ok := tagToField[key]; ok {
			field := v.Field(fieldIdx)
			setFieldValue(field, val)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading alink config: %w", err)
	}

	return config, nil
}

// SaveAlink updates specific keys in the Alink configuration file
func (s *ServiceConfig) SaveAlink(config *models.AlinkConfig) error {
	// Read existing file line by line
	existingLines, err := readLines(s.AlinkPath)
	if err != nil {
		return err
	}

	// Create a map of keys to update from the struct
	// We iterate over all fields in the struct and add them to updates map
	updates := make(map[string]string)
	v := reflect.ValueOf(config).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("conf")
		if tag == "" {
			continue
		}

		val := getFieldValueString(v.Field(i))
		updates[tag] = val
	}

	var newLines []string
	updatedKeys := make(map[string]bool)

	for _, line := range existingLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			newLines = append(newLines, line)
			continue
		}

		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			newLines = append(newLines, line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		// normalizedKey := key // Case sensitivity? Alink seems case sensitive (lowercase)

		if newVal, ok := updates[key]; ok {
			newLines = append(newLines, key+"="+newVal)
			updatedKeys[key] = true
		} else {
			newLines = append(newLines, line)
		}
	}

	// Append missing keys
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("conf")
		if tag != "" && !updatedKeys[tag] {
			// Get value again
			// Note: This relies on iteration order which is fixed for structs
			val := updates[tag]
			newLines = append(newLines, tag+"="+val)
		}
	}

	// Write back to file
	tmpPath := s.AlinkPath + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(strings.Join(newLines, "\n")+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to write temp alink config: %w", err)
	}

	if err := os.Rename(tmpPath, s.AlinkPath); err != nil {
		return fmt.Errorf("failed to replace alink config: %w", err)
	}

	return nil
}

func setFieldValue(field reflect.Value, val string) {
	switch field.Kind() {
	case reflect.String:
		field.SetString(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v, err := strconv.Atoi(val); err == nil {
			field.SetInt(int64(v))
		}
	case reflect.Bool:
		field.SetBool(val == "1" || val == "true")
	case reflect.Float64, reflect.Float32:
		if v, err := strconv.ParseFloat(val, 64); err == nil {
			field.SetFloat(v)
		}
	}
}

func getFieldValueString(field reflect.Value) string {
	switch field.Kind() {
	case reflect.String:
		return field.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(field.Int(), 10)
	case reflect.Bool:
		if field.Bool() {
			return "1"
		}
		return "0"
	case reflect.Float64, reflect.Float32:
		return strconv.FormatFloat(field.Float(), 'f', -1, 64)
	}
	return ""
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		// If file doesn't exist, return empty list
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
