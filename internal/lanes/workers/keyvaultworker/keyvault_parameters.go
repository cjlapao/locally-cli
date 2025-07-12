package keyvaultworker

import (
	"github.com/cjlapao/locally-cli/internal/environment"
)

type KeyvaultParameters struct {
	Name         string `json:"name,omitempty" yaml:"name,omitempty"`
	KeyvaultUrl  string `json:"keyvaultUrl,omitempty" yaml:"keyvaultUrl,omitempty"`
	Base64Decode bool   `json:"base64Decode,omitempty" yaml:"base64Decode,omitempty"`
}

func (c *KeyvaultParameters) Validate() bool {
	return c.KeyvaultUrl != ""
}

func (c *KeyvaultParameters) Decode() {
	env := environment.Get()

	c.KeyvaultUrl = env.Replace(c.KeyvaultUrl)
	c.Name = env.Replace(c.Name)
}
