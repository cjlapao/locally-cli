package models

import (
	"fmt"
	"strings"
	"time"
)

type Claim struct {
	ID            string        `json:"id" yaml:"id"`
	TenantID      string        `json:"tenant_id" yaml:"tenant_id"`
	Slug          string        `json:"slug" yaml:"slug"`
	Module        string        `json:"module" yaml:"module"`
	Service       string        `json:"service" yaml:"service"`
	Action        AccessLevel   `json:"action" yaml:"action"`
	CreatedAt     time.Time     `json:"created_at" yaml:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" yaml:"updated_at"`
	SecurityLevel SecurityLevel `json:"level" yaml:"level"`
	Matched       bool          `json:"-" yaml:"-"`
}

func (c *Claim) GetModule() string {
	if c.Module == "*" {
		return "*"
	}
	return c.Module
}

func (c *Claim) GetService() string {
	if c.Service == "*" {
		return "*"
	}
	return c.Service
}

func (c *Claim) GetAction() string {
	if c.Action == AccessLevelAll {
		return "*"
	}
	return string(c.Action)
}

func (c *Claim) GetSlug() string {
	return GetClaimName(c)
}

// CanAccess checks if this claim can access the required claim
func (c *Claim) CanAccess(required *Claim) bool {
	// Check if services match (or if either is wildcard)
	if c.Service != "*" && required.Service != "*" && c.Service != required.Service {
		return false
	}

	// Check if modules match (or if either is wildcard)
	if c.Module != "*" && required.Module != "*" && c.Module != required.Module {
		return false
	}

	// Check if action can access required action
	return c.Action.CanAccess(required.Action)
}

// ParseClaim parses a claim from a slug in the format "service::module::action"
func ParseClaim(slug string) (*Claim, error) {
	// Split the slug by "::"
	parts := strings.Split(slug, "::")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid claim format: expected 'service::module::action', got '%s'", slug)
	}

	service := parts[0]
	module := parts[1]
	actionStr := parts[2]

	// Validate action
	action := AccessLevel(actionStr)
	switch action {
	case AccessLevelRead, AccessLevelWrite, AccessLevelDelete, AccessLevelAll, AccessLevelNone,
		AccessLevelUpdate, AccessLevelCreate, AccessLevelView, AccessLevelApprove, AccessLevelReject,
		AccessLevelCancel, AccessLevelSuspend, AccessLevelResume, AccessLevelReset, AccessLevelUnlock, AccessLevelLock:
		// Valid action
	default:
		return nil, fmt.Errorf("invalid action '%s' in claim slug '%s'", actionStr, slug)
	}

	// Create the claim
	claim := &Claim{
		Service: service,
		Module:  module,
		Action:  action,
		Slug:    slug,
	}

	return claim, nil
}

func GetClaimName(claim *Claim) string {
	return fmt.Sprintf("%s::%s::%s", claim.GetService(), claim.GetModule(), claim.GetAction())
}
