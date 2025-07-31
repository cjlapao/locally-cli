package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Server represents the HTTP server
type Server struct {
	server              *http.Server
	port                int
	hostname            string
	prefix              string
	handler             *Handler
	router              *mux.Router
	middlewareChain     *MiddlewareChain
	routeGroups         []RouteGroup
	authMiddleware      PreMiddleware
	superUserMiddleware PreMiddleware
	roleMiddleware      PreMiddleware
	claimMiddleware     PreMiddleware
}

// Config represents the server configuration
type Config struct {
	Port                int
	Hostname            string
	Prefix              string
	AuthMiddleware      PreMiddleware
	SuperUserMiddleware PreMiddleware
	RoleMiddleware      PreMiddleware
	ClaimMiddleware     PreMiddleware
	CORSConfig          *CORSConfig // Optional CORS configuration
}

// NewServer creates a new HTTP server
func NewServer(cfg Config, handler *Handler) *Server {
	appCfg := config.GetInstance().Get()

	// Create middleware chain with default middlewares
	middlewareChain := NewMiddlewareChain()
	middlewareChain.AddPreMiddleware(CORSMiddleware(readCorsConfigFromConfiguration(appCfg)))
	middlewareChain.AddPreMiddleware(RequestIDMiddleware())
	middlewareChain.AddPreMiddleware(RequestLoggingMiddleware())
	middlewareChain.AddPostMiddleware(ResponseLoggingMiddleware())

	return &Server{
		handler:             handler,
		port:                cfg.Port,
		hostname:            cfg.Hostname,
		prefix:              cfg.Prefix,
		router:              mux.NewRouter(),
		middlewareChain:     middlewareChain,
		routeGroups:         make([]RouteGroup, 0),
		authMiddleware:      cfg.AuthMiddleware,
		superUserMiddleware: cfg.SuperUserMiddleware,
		roleMiddleware:      cfg.RoleMiddleware,
		claimMiddleware:     cfg.ClaimMiddleware,
	}
}

// RegisterRoutes registers routes from a RouteRegistrar
func (s *Server) RegisterRoutes(registrar RouteRegistrar) {
	routes := registrar.Routes()
	for _, route := range routes {
		s.registerRoute(route)
	}
}

// RegisterRouteGroup registers a group of routes with common prefix and middleware
func (s *Server) RegisterRouteGroup(group RouteGroup) {
	s.routeGroups = append(s.routeGroups, group)
}

// AddPreMiddleware adds a pre-middleware to the global chain
func (s *Server) AddPreMiddleware(middleware PreMiddleware) {
	s.middlewareChain.AddPreMiddleware(middleware)
}

// AddPostMiddleware adds a post-middleware to the global chain
func (s *Server) AddPostMiddleware(middleware PostMiddleware) {
	s.middlewareChain.AddPostMiddleware(middleware)
}

// registerRoute registers a single route with the server
func (s *Server) registerRoute(route Route) {
	handler := route.Handler

	// Apply route-specific middleware first
	for i := len(route.Middleware) - 1; i >= 0; i-- {
		handler = route.Middleware[i](handler)
	}

	// Create a custom middleware chain for this route
	routeChain := NewMiddlewareChain()

	// Add global pre-middlewares
	for _, middleware := range s.middlewareChain.preMiddlewares {
		routeChain.AddPreMiddleware(middleware)
	}

	// Add auth middleware if required
	if route.AuthRequired && s.authMiddleware != nil {
		routeChain.AddPreMiddleware(s.authMiddleware)
	}

	// Add super user middleware if required
	if route.SuperUserRequired && s.superUserMiddleware != nil {
		routeChain.AddPreMiddleware(s.superUserMiddleware)
	}

	// Add role middleware if required
	if len(route.Roles) > 0 {
		routeChain.AddPreMiddleware(NewRequireRolePreMiddleware(route.Roles))
	}

	// Add claim middleware if required
	if len(route.Claims) > 0 {
		routeChain.AddPreMiddleware(NewRequireClaimPreMiddleware(route.Claims))
	}

	// Add global post-middlewares
	for _, middleware := range s.middlewareChain.postMiddlewares {
		routeChain.AddPostMiddleware(middleware)
	}

	// Execute the middleware chain
	finalHandler := routeChain.Execute(handler)

	if s.prefix != "" {
		if !strings.HasPrefix(route.Path, "/") {
			route.Path = "/" + route.Path
		}
		route.Path = s.prefix + route.Path

		s.router.HandleFunc(route.Path, finalHandler).Methods(route.Method)

		if route.Method != http.MethodOptions {
			s.router.HandleFunc(route.Path, s.createOptionsHandler()).Methods(http.MethodOptions)
		}
	} else {
		s.router.HandleFunc(route.Path, finalHandler).Methods(route.Method)

		if route.Method != http.MethodOptions {
			s.router.HandleFunc(route.Path, s.createOptionsHandler()).Methods(http.MethodOptions)
		}
	}

	logging.WithFields(logrus.Fields{
		"method":              route.Method,
		"path":                route.Path,
		"description":         route.Description,
		"auth_required":       route.AuthRequired,
		"super_user_required": route.SuperUserRequired,
	}).Info("Registered route")
}

// createOptionsHandler creates a handler for OPTIONS requests
func (s *Server) createOptionsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCfg := config.GetInstance().Get()
		// Create a minimal middleware chain for OPTIONS (just CORS)
		optionsChain := NewMiddlewareChain()
		optionsChain.AddPreMiddleware(CORSMiddleware(readCorsConfigFromConfiguration(appCfg)))

		// Create a simple handler that returns 204 No Content
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		// Execute the chain
		optionsChain.Execute(handler)(w, r)
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	for _, group := range s.routeGroups {
		subrouter := s.router.PathPrefix(group.Prefix).Subrouter()

		for _, route := range group.Routes {
			handler := route.Handler

			// Apply route-specific middleware first
			for i := len(route.Middleware) - 1; i >= 0; i-- {
				handler = route.Middleware[i](handler)
			}

			// Apply group middleware
			for i := len(group.Middleware) - 1; i >= 0; i-- {
				handler = group.Middleware[i](handler)
			}

			// Create a custom middleware chain for this route
			routeChain := NewMiddlewareChain()

			// Add global pre-middlewares
			for _, middleware := range s.middlewareChain.preMiddlewares {
				routeChain.AddPreMiddleware(middleware)
			}

			// Add auth middleware if required
			if route.AuthRequired && s.authMiddleware != nil {
				routeChain.AddPreMiddleware(s.authMiddleware)
			}

			// Add super user middleware if required
			if route.SuperUserRequired && s.superUserMiddleware != nil {
				routeChain.AddPreMiddleware(s.superUserMiddleware)
			}

			// Add role middleware if required
			if len(route.Roles) > 0 {
				routeChain.AddPreMiddleware(s.roleMiddleware)
			}

			// Add claim middleware if required
			if len(route.Claims) > 0 {
				routeChain.AddPreMiddleware(s.claimMiddleware)
			}

			// Add global post-middlewares
			for _, middleware := range s.middlewareChain.postMiddlewares {
				routeChain.AddPostMiddleware(middleware)
			}

			// Execute the middleware chain
			finalHandler := routeChain.Execute(handler)

			subrouter.HandleFunc(route.Path, finalHandler).Methods(route.Method)

			if route.Method != http.MethodOptions {
				subrouter.HandleFunc(route.Path, s.createOptionsHandler()).Methods(http.MethodOptions)
			}

			logging.WithFields(logrus.Fields{
				"method":              route.Method,
				"path":                group.Prefix + route.Path,
				"description":         route.Description,
				"auth_required":       route.AuthRequired,
				"super_user_required": route.SuperUserRequired,
			}).Info("Registered group route")
		}
	}

	addr := fmt.Sprintf("%s:%d", s.hostname, s.port)
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	logging.WithField("address", addr).Info("Starting server")
	return s.server.ListenAndServe()
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

func readCorsConfigFromConfiguration(cfg *config.Config) CORSConfig {
	corsConfig := DefaultCORSConfig()
	corsAllowOrigins := cfg.Get(config.CorsAllowOriginsKey).GetString()
	corsAllowMethods := cfg.Get(config.CorsAllowMethodsKey).GetString()
	corsAllowHeaders := cfg.Get(config.CorsAllowHeadersKey).GetString()
	corsExposeHeaders := cfg.Get(config.CorsExposeHeadersKey).GetString()
	if corsAllowOrigins != "" {
		configCorsOrigins := strings.Split(corsAllowOrigins, ",")
		corsConfig.AllowOrigins = make([]string, 0)
		for _, origin := range configCorsOrigins {
			corsConfig.AllowOrigins = append(corsConfig.AllowOrigins, strings.TrimSpace(origin))
		}
	}
	if corsAllowMethods != "" {
		configCorsMethods := strings.Split(corsAllowMethods, ",")
		corsConfig.AllowMethods = make([]string, 0)
		for _, method := range configCorsMethods {
			corsConfig.AllowMethods = append(corsConfig.AllowMethods, strings.TrimSpace(method))
		}
	}
	if corsAllowHeaders != "" {
		configCorsHeaders := strings.Split(corsAllowHeaders, ",")
		corsConfig.AllowHeaders = make([]string, 0)
		for _, header := range configCorsHeaders {
			corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, strings.TrimSpace(header))
		}
	}
	if corsExposeHeaders != "" {
		configCorsExposeHeaders := strings.Split(corsExposeHeaders, ",")
		corsConfig.ExposeHeaders = make([]string, 0)
		for _, header := range configCorsExposeHeaders {
			corsConfig.ExposeHeaders = append(corsConfig.ExposeHeaders, strings.TrimSpace(header))
		}
	}

	// Add X-Tenant-ID to the allow headers if it's not already there, this is a
	// special header that is used to identify the tenant in the request
	if !strings.Contains(strings.Join(corsConfig.AllowHeaders, ","), "X-Tenant-ID") {
		corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "X-Tenant-ID")
	}
	return corsConfig
}
