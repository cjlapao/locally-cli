package azure_cli

import "regexp"

func ValidateStorageAccountName(name string) bool {
	if name == "" {
		return false
	}

	if len(name) < 3 || len(name) > 24 {
		return false
	}

	rx, err := regexp.MatchString(`[^a-z\d]`, name)
	if err != nil {
		return false
	}

	return !rx
}
