package curlworker

// import (
// 	"github.com/cjlapao/locally-cli/internal/environment"
// )

// type CurlParameters struct {
// 	Host             string            `json:"host,omitempty" yaml:"host,omitempty"`
// 	Verb             string            `json:"verb,omitempty" yaml:"verb,omitempty"`
// 	Headers          map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
// 	Content          *CurlContent      `json:"content,omitempty" yaml:"content,omitempty"`
// 	RetryCount       int               `json:"retryCount,omitempty" yaml:"retryCount,omitempty"`
// 	WaitForInSeconds int               `json:"waitFor,omitempty" yaml:"waitFor,omitempty"`
// }

// func (c *CurlParameters) Validate() bool {
// 	if c.Host == "" {
// 		return false
// 	}
// 	if c.Verb == "" {
// 		c.Verb = "GET"
// 	}

// 	return true
// }

// type CurlContent struct {
// 	ContentType string            `json:"contentType,omitempty" yaml:"contentType,omitempty"`
// 	UrlEncoded  map[string]string `json:"urlEncoded,omitempty" yaml:"urlEncoded,omitempty"`
// 	Json        string            `json:"json,omitempty" yaml:"json,omitempty"`
// }

// func (c *CurlParameters) Decode() {
// 	env := environment.GetInstance()

// 	c.Host = env.Replace(c.Host)
// 	for key, value := range c.Headers {
// 		c.Headers[key] = env.Replace(value)
// 	}

// 	if c.Content != nil {
// 		if c.Content.ContentType != "" {
// 			c.Content.ContentType = env.Replace(c.Content.ContentType)
// 		}

// 		if c.Content.ContentType != "" {
// 			c.Content.Json = env.Replace(c.Content.Json)
// 		}

// 		for key, value := range c.Content.UrlEncoded {
// 			c.Content.UrlEncoded[key] = env.Replace(value)
// 		}
// 	}
// }
