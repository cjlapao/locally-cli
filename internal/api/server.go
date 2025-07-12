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
	server         *http.Server
	port           int
	hostname       string
	prefix         string
	handler        *Handler
	router         *mux.Router
	middleware     []Middleware
	routeGroups    []RouteGroup
	authMiddleware Middleware
}

// Config represents the server configuration
type Config struct {
	Port           int
	Hostname       string
	Prefix         string
	AuthMiddleware Middleware
	CORSConfig     *CORSConfig // Optional CORS configuration
}

// NewServer creates a new HTTP server
func NewServer(cfg Config, handler *Handler) *Server {
	corsConfig := DefaultCORSConfig()
	if cfg.CORSConfig != nil {
		corsConfig = *cfg.CORSConfig
	}
	appCfg := config.GetInstance().Get()
	corsAllowOrigins := appCfg.Get(config.CorsAllowOriginsKey).GetString()
	corsAllowMethods := appCfg.Get(config.CorsAllowMethodsKey).GetString()
	corsAllowHeaders := appCfg.Get(config.CorsAllowHeadersKey).GetString()
	corsExposeHeaders := appCfg.Get(config.CorsExposeHeadersKey).GetString()
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

	return &Server{
		handler:        handler,
		port:           cfg.Port,
		hostname:       cfg.Hostname,
		prefix:         cfg.Prefix,
		router:         mux.NewRouter(),
		middleware:     []Middleware{RequestIDMiddleware, CORSMiddleware(corsConfig), LoggingMiddleware},
		routeGroups:    make([]RouteGroup, 0),
		authMiddleware: cfg.AuthMiddleware,
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

// registerRoute registers a single route with the server
func (s *Server) registerRoute(route Route) {
	handler := route.Handler

	for i := len(route.Middleware) - 1; i >= 0; i-- {
		handler = route.Middleware[i](handler)
	}

	if route.AuthRequired && s.authMiddleware != nil {
		handler = s.authMiddleware(handler)
	}

	for i := len(s.middleware) - 1; i >= 0; i-- {
		handler = s.middleware[i](handler)
	}

	if s.prefix != "" {
		if !strings.HasPrefix(route.Path, "/") {
			route.Path = "/" + route.Path
		}
		route.Path = s.prefix + route.Path

		s.router.HandleFunc(route.Path, handler).Methods(route.Method)

		if route.Method != http.MethodOptions {
			s.router.HandleFunc(route.Path, s.createOptionsHandler()).Methods(http.MethodOptions)
		}
	} else {
		s.router.HandleFunc(route.Path, handler).Methods(route.Method)

		if route.Method != http.MethodOptions {
			s.router.HandleFunc(route.Path, s.createOptionsHandler()).Methods(http.MethodOptions)
		}
	}

	logging.WithFields(logrus.Fields{
		"method":        route.Method,
		"path":          route.Path,
		"description":   route.Description,
		"auth_required": route.AuthRequired,
	}).Info("Registered route")
}

// createOptionsHandler creates a handler for OPTIONS requests
func (s *Server) createOptionsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Apply only the global middleware chain (which includes CORS)
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// The CORS middleware will handle the OPTIONS request
			// and return early, so this should never be reached
			w.WriteHeader(http.StatusNoContent)
		})

		// Apply global middleware
		for i := len(s.middleware) - 1; i >= 0; i-- {
			handler = s.middleware[i](handler)
		}

		handler(w, r)
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	for _, group := range s.routeGroups {
		subrouter := s.router.PathPrefix(group.Prefix).Subrouter()

		for _, route := range group.Routes {
			handler := route.Handler
			for i := len(route.Middleware) - 1; i >= 0; i-- {
				handler = route.Middleware[i](handler)
			}

			for i := len(group.Middleware) - 1; i >= 0; i-- {
				handler = group.Middleware[i](handler)
			}

			if route.AuthRequired && s.authMiddleware != nil {
				handler = s.authMiddleware(handler)
			}

			for i := len(s.middleware) - 1; i >= 0; i-- {
				handler = s.middleware[i](handler)
			}

			subrouter.HandleFunc(route.Path, handler).Methods(route.Method)

			if route.Method != http.MethodOptions {
				subrouter.HandleFunc(route.Path, s.createOptionsHandler()).Methods(http.MethodOptions)
			}

			logging.WithFields(logrus.Fields{
				"method":        route.Method,
				"path":          group.Prefix + route.Path,
				"description":   route.Description,
				"auth_required": route.AuthRequired,
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

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	logging.Info("Stopping server...")
	return s.server.Shutdown(ctx)
}
