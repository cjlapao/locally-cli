package common

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/cjlapao/common-go/helper"
)

func ExtractUri(uri string) string {
	parts := strings.Split(uri, ".")
	if len(parts) == 0 {
		return ""
	}

	return parts[0]
}

func VerifyCommand(command string) string {
	if strings.HasPrefix(command, helper.FlagPrefix) {
		return ""
	}

	return command
}

func GetHostFromUrl(urlString string) string {
	parsedUrl, err := url.Parse(urlString)
	if err != nil {
		return urlString
	}

	if parsedUrl.Host == "" {
		return urlString
	}

	return parsedUrl.Host
}

func GetExeDirectoryPath() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	return filepath.Dir(ex)
}

func EncodeName(name string) string {
	folderName := strings.ReplaceAll(name, "\\", "/")
	folderName = strings.ReplaceAll(folderName, " ", "_")
	folderName = strings.ReplaceAll(folderName, ".", "_")

	return strings.ToLower(folderName)
}
