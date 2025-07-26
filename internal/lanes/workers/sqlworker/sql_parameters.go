package sqlworker

import (
	"github.com/cjlapao/locally-cli/internal/environment"
)

type SqlParameters struct {
	ConnectionString string `json:"connectionString,omitempty" yaml:"connectionString,omitempty"`
	Query            string `json:"query,omitempty" yaml:"query,omitempty"`
}

func (c *SqlParameters) Validate() bool {
	return c.ConnectionString != "" && c.Query != ""
}

func (c *SqlParameters) Decode() {
	env := environment.GetInstance()

	c.ConnectionString = env.Replace(c.ConnectionString)
	c.Query = env.Replace(c.Query)
}
