// Package types provides the types for the API
package types

import (
	"net/http"

	"github.com/cjlapao/locally-cli/pkg/models"
)

// Route represents an API route
type Route struct {
	Method              string
	Path                string
	Handler             http.HandlerFunc
	Middleware          []func(http.HandlerFunc) http.HandlerFunc // Legacy middleware for backward compatibility
	Description         string
	SecurityRequirement *SecurityRequirement
}

// RouteGroup represents a group of related routes
type RouteGroup struct {
	Prefix     string
	Routes     []Route
	Middleware []func(http.HandlerFunc) http.HandlerFunc // Legacy middleware for backward compatibility
}

// RouteRegistrar interface for modules that provide routes
type RouteRegistrar interface {
	Routes() []Route
}

type AuthorizationHeaderType string

const (
	AuthorizationHeaderTypeBearer AuthorizationHeaderType = "Bearer"
	AuthorizationHeaderTypeBasic  AuthorizationHeaderType = "Basic"
	AuthorizationHeaderTypeApiKey AuthorizationHeaderType = "ApiKey"
	AuthorizationHeaderTypeNone   AuthorizationHeaderType = "None"
)

type AuthorizationHeader struct {
	AuthorizationType AuthorizationHeaderType
	Token             string
}

type SecurityRequirement struct {
	SecurityLevel models.ApiKeySecurityLevel
	Claims        *SecurityRequirementClaims
	Roles         *SecurityRequirementRoles
}

type SecurityRequirementRelation string

const (
	SecurityRequirementRelationAnd SecurityRequirementRelation = "and"
	SecurityRequirementRelationOr  SecurityRequirementRelation = "or"
)

type SecurityRequirementClaims struct {
	Relation SecurityRequirementRelation
	Items    []models.Claim
}

type SecurityRequirementRoles struct {
	Relation SecurityRequirementRelation
	Items    []models.Role
}
