package tools

import "encoding/base64"

type Base64Tool struct{}

func (b Base64Tool) Encode(value string) string {
	return base64.StdEncoding.EncodeToString([]byte(value))
}

func (b Base64Tool) Decode(value string) (string, error) {
	bValue, err := base64.StdEncoding.DecodeString(value)

	return string(bValue), err
}
