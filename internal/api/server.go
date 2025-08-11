package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/cjlapao/locally-cli/internal/api/middleware"
	"github.com/cjlapao/locally-cli/internal/api/types"
	auth_interfaces "github.com/cjlapao/locally-cli/internal/auth/interfaces"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/cjlapao/locally-cli/pkg/models"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Server represents the HTTP server
type Server struct {
	server          *http.Server
	port            int
	hostname        string
	prefix          string
	router          *mux.Router
	middlewareChain *middleware.MiddlewareChain
	routeGroups     []types.RouteGroup
	authService     auth_interfaces.AuthServiceInterface
}

// Config represents the server configuration
type Config struct {
	Port        int
	Hostname    string
	Prefix      string
	AuthService auth_interfaces.AuthServiceInterface
	CORSConfig  *middleware.CORSConfig // Optional CORS configuration
}

// NewServer creates a new HTTP server
func NewServer(cfg Config) *Server {
	appCfg := config.GetInstance().Get()

	// Create middleware chain with default middlewares
	middlewareChain := middleware.NewMiddlewareChain()
	middlewareChain.AddPreMiddleware(middleware.CORSMiddleware(readCorsConfigFromConfiguration(appCfg)))
	middlewareChain.AddPreMiddleware(middleware.RequestInformationMiddleware())
	middlewareChain.AddPreMiddleware(middleware.RequestLoggingMiddleware())
	middlewareChain.AddPostMiddleware(middleware.ResponseLoggingMiddleware())

	return &Server{
		port:            cfg.Port,
		hostname:        cfg.Hostname,
		prefix:          cfg.Prefix,
		router:          mux.NewRouter(),
		middlewareChain: middlewareChain,
		routeGroups:     make([]types.RouteGroup, 0),
		authService:     cfg.AuthService,
	}
}

// RegisterRoutes registers routes from a RouteRegistrar
func (s *Server) RegisterRoutes(registrar types.RouteRegistrar) {
	routes := registrar.Routes()
	for _, route := range routes {
		s.registerRoute(route)
	}
}

// RegisterRouteGroup registers a group of routes with common prefix and middleware
func (s *Server) RegisterRouteGroup(group types.RouteGroup) {
	s.routeGroups = append(s.routeGroups, group)
}

// AddPreMiddleware adds a pre-middleware to the global chain
func (s *Server) AddPreMiddleware(middleware middleware.PreMiddleware) {
	s.middlewareChain.AddPreMiddleware(middleware)
}

// AddPostMiddleware adds a post-middleware to the global chain
func (s *Server) AddPostMiddleware(middleware middleware.PostMiddleware) {
	s.middlewareChain.AddPostMiddleware(middleware)
}

// registerRoute registers a single route with the server
func (s *Server) registerRoute(route types.Route) {
	handler := route.Handler

	// Apply route-specific middleware first
	for i := len(route.Middleware) - 1; i >= 0; i-- {
		handler = route.Middleware[i](handler)
	}

	// Create a custom middleware chain for this route
	routeChain := middleware.NewMiddlewareChain()

	// Add global pre-middlewares
	for _, middleware := range s.middlewareChain.PreMiddlewares {
		routeChain.AddPreMiddleware(middleware)
	}

	// Add new auth middleware if required
	routeChain.AddPreMiddleware(middleware.NewAuthorizationPreMiddleware(s.authService, &route))

	// Add global post-middlewares
	for _, middleware := range s.middlewareChain.PostMiddlewares {
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

	securityLevel := models.ApiKeySecurityLevelNone
	if route.SecurityRequirement != nil {
		securityLevel = route.SecurityRequirement.SecurityLevel
	}

	logging.WithFields(logrus.Fields{
		"method":         route.Method,
		"path":           route.Path,
		"description":    route.Description,
		"security_level": securityLevel,
	}).Info("Registered route")
}

// createOptionsHandler creates a handler for OPTIONS requests
func (s *Server) createOptionsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCfg := config.GetInstance().Get()
		// Create a minimal middleware chain for OPTIONS (just CORS)
		optionsChain := middleware.NewMiddlewareChain()
		optionsChain.AddPreMiddleware(middleware.CORSMiddleware(readCorsConfigFromConfiguration(appCfg)))

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
			routeChain := middleware.NewMiddlewareChain()

			// Add global pre-middlewares
			for _, middleware := range s.middlewareChain.PreMiddlewares {
				routeChain.AddPreMiddleware(middleware)
			}

			routeChain.AddPreMiddleware(middleware.NewAuthorizationPreMiddleware(s.authService, &route))

			// Add global post-middlewares
			for _, middleware := range s.middlewareChain.PostMiddlewares {
				routeChain.AddPostMiddleware(middleware)
			}

			// Execute the middleware chain
			finalHandler := routeChain.Execute(handler)

			subrouter.HandleFunc(route.Path, finalHandler).Methods(route.Method)

			if route.Method != http.MethodOptions {
				subrouter.HandleFunc(route.Path, s.createOptionsHandler()).Methods(http.MethodOptions)
			}

			securityLevel := models.ApiKeySecurityLevelNone
			if route.SecurityRequirement != nil {
				securityLevel = route.SecurityRequirement.SecurityLevel
			}

			logging.WithFields(logrus.Fields{
				"method":         route.Method,
				"path":           group.Prefix + route.Path,
				"description":    route.Description,
				"security_level": securityLevel,
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

func readCorsConfigFromConfiguration(cfg *config.Config) middleware.CORSConfig {
	corsConfig := middleware.DefaultCORSConfig()
	systemHeadersToAllow := []string{
		"X-Tenant-ID",
		"Content-Type",
		"Host",
		"User-Agent",
		"Accept",
		"Accept-Encoding",
		"Accept-Language",
		"Origin",
		"Referer",
		"Postman-Token",
		"Cache-Control",
		"Pragma",
		"Connection",
		"Upgrade-Insecure-Requests",
		"Sec-Fetch-Dest",
		"Sec-Fetch-Mode",
		"Sec-Fetch-Site",
		"Content-Length",
		"Sec-Fetch-User",
	}
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

	// if the headers in the config are * then we do not need to add the special ones as all are allowed
	if strings.Contains(strings.Join(corsConfig.AllowHeaders, ","), "*") {
		corsConfig.AllowHeaders = []string{"*"}
	} else {
		for _, header := range systemHeadersToAllow {
			if !strings.Contains(strings.Join(corsConfig.AllowHeaders, ","), header) {
				if !strings.Contains(strings.Join(corsConfig.AllowHeaders, ","), "*") {
					corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, header)
				}
			}
		}
	}
	return corsConfig
}
