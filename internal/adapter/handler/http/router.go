package http

import (
	"strings"

	"leeta/internal/adapter/config"

	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// Router is a wrapper for HTTP router
type Router struct {
	chi.Router
}

// NewRouter creates a new HTTP router
func NewRouter(
	config *config.ServerConfiguration,
	logger *zap.Logger,
	pingHandler PingHandler,
	locationHandler LocationHandler,
) (*Router, error) {

	// CORS
	corsConfig := cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}

	allowedOrigins := config.HttpAllowedOrigins
	if allowedOrigins != "" {
		originsList := strings.Split(config.HttpAllowedOrigins, ",")
		corsConfig.AllowedOrigins = originsList
	}

	router := chi.NewRouter()
	router.Use(cors.Handler(corsConfig))

	// Logger
	router.Use(requestLogger)
	router.Use(middleware.Recoverer)

	// Swagger
	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("0.0.0.0:"+config.HttpPort+"/swagger/doc.json"), //The url pointing to API definition
	))

	// v1
	router.Route("/v1", func(r chi.Router) {

		// Ping
		r.Route("/health", func(r chi.Router) {
			r.Get("/", pingHandler.PingGet)
			r.Post("/", pingHandler.PingPost)
		})

		// Location
		r.Route("/locations", func(r chi.Router) {
			r.Post("/", locationHandler.RegisterLocation)
			r.Get("/{name}", locationHandler.GetLocation)
			r.Delete("/{name}", locationHandler.DeleteLocation)
			r.Get("/", locationHandler.ListLocations)
			r.Get("/nearest", locationHandler.GetNearestLocation)
		})

	})

	return &Router{
		router,
	}, nil
}
