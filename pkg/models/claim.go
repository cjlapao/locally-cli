package models

import (
	"fmt"
	"time"

	"github.com/cjlapao/locally-cli/pkg/utils"
)

type ClaimAction string

const (
	ClaimActionRead   ClaimAction = "read"
	ClaimActionWrite  ClaimAction = "write"
	ClaimActionDelete ClaimAction = "delete"
	ClaimActionAll    ClaimAction = "*"
	ClaimActionNone   ClaimAction = "none"
	ClaimActionUpdate ClaimAction = "update"
	ClaimActionCreate ClaimAction = "view"
)

type Claim struct {
	ID        string      `json:"id" yaml:"id"`
	Slug      string      `json:"slug" yaml:"slug"`
	Module    string      `json:"module" yaml:"module"`
	Service   string      `json:"service" yaml:"service"`
	Action    ClaimAction `json:"action" yaml:"action"`
	CreatedAt time.Time   `json:"created_at" yaml:"created_at"`
	UpdatedAt time.Time   `json:"updated_at" yaml:"updated_at"`
	Matched   bool        `json:"-" yaml:"-"`
}

func (c *Claim) GetModule() string {
	if c.Module == "*" {
		return "all"
	}
	return c.Module
}

func (c *Claim) GetService() string {
	if c.Service == "*" {
		return "all"
	}
	return c.Service
}

func (c *Claim) GetAction() string {
	if c.Action == ClaimActionAll {
		return "all"
	}
	return string(c.Action)
}

func (c *Claim) GetName() string {
	return fmt.Sprintf("%s::%s::%s", c.GetService(), c.GetModule(), c.GetAction())
}

func (c *Claim) GetSlug() string {
	return utils.Slugify(c.GetName())
}
