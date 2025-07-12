package api

import "net/http"

// Route represents an API route
type Route struct {
	Method       string
	Path         string
	Handler      http.HandlerFunc
	Middleware   []Middleware
	Description  string
	AuthRequired bool
}

// RouteGroup represents a group of related routes
type RouteGroup struct {
	Prefix     string
	Routes     []Route
	Middleware []Middleware
}

// Middleware represents a middleware function
type Middleware func(http.HandlerFunc) http.HandlerFunc

// RouteRegistrar interface for modules that provide routes
type RouteRegistrar interface {
	Routes() []Route
}
