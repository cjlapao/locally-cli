package api

import (
	"net/http"

	"github.com/cjlapao/locally-cli/pkg/models"
)

// Route represents an API route
type Route struct {
	Method        string
	Path          string
	Handler       http.HandlerFunc
	Middleware    []func(http.HandlerFunc) http.HandlerFunc // Legacy middleware for backward compatibility
	Description   string
	SecurityLevel models.ApiKeySecurityLevel
	Claims        []models.Claim
	Roles         []models.Role
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
