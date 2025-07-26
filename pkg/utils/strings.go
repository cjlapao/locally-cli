package utils

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
)

// ObfuscateString obfuscates a password string for display purposes
// If the string is 5 characters or less, returns "***"
// If the string is longer than 5 characters, returns first char + "***" + last char
func ObfuscateString(password string) string {
	if len(password) <= 5 {
		return "***"
	}

	return string(password[0]) + "***" + string(password[len(password)-1])
}

func Slugify(input string) string {
	// replace anything that is not an alphanumeric, a underscore, or a dash with a dash using regex
	re := regexp.MustCompile("[^a-zA-Z0-9_-]+")
	return strings.ToLower(re.ReplaceAllString(input, "-"))
}

// StringToMap Function that takes a string and marshals it to a map[string]interface{}
func StringToMap(input string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(input), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func StringToSlice(input string) ([]string, error) {
	var result []string
	err := json.Unmarshal([]byte(input), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func StringToObject[T any](input string) (T, error) {
	var result T
	err := json.Unmarshal([]byte(input), &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func ObjectToJson[T any](input T) (string, error) {
	json, err := json.Marshal(input)
	// making it as small as possible so remove all the whitespace
	json = bytes.TrimSpace(json)
	if err != nil {
		return "", err
	}
	return string(json), nil
}
