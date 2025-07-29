package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	_ "leeta/docs"
	"leeta/internal/adapter/config"
	httpHandler "leeta/internal/adapter/handler/http"
	"leeta/internal/adapter/logger"
	"leeta/internal/adapter/storage/postgres"
	"leeta/internal/adapter/storage/postgres/repository"
	"leeta/internal/core/service"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// @title			Leeta Golang Exercise
// @version		1.0
// @description	Find nearest places to a given location
//
// @contact.name	Emmanuel Jonathan
// @contact.url	https://github.com/emmrys-jay
// @contact.email	jonathanemma121@gmail.com
//
// @host			localhost:8081
// @BasePath		/v1
// @schemes		http https
func main() {
	// Load environment variables
	config := config.Setup()

	// Set logger
	l := logger.Get()

	l.Info("Starting the application",
		zap.String("app", config.App.Name),
		zap.String("env", config.App.Env))

	ctx := context.Background()

	// Init database
	db, err := postgres.New(ctx, &config.Database)
	if err != nil {
		l.Error("Error initializing database connection", zap.Error(err))
		os.Exit(1)
	}
	defer db.Close()

	l.Info("Successfully connected to the database",
		zap.String("db", config.Database.Protocol))

	// Migrate postgres database
	err = db.Migrate()
	if err != nil {
		l.Error("Error migrating database", zap.Error(err))
		os.Exit(1)
	}

	l.Info("Successfully migrated the database")

	// Dependency injection
	// Ping
	pingRepo := repository.NewPingRepository(db)
	pingService := service.NewPingService(pingRepo)
	pingHandler := httpHandler.NewPingHandler(pingService, validator.New())

	// Location
	locationRepo := repository.NewLocationRepository(db)
	locationService := service.NewLocationService(locationRepo)
	locationHandler := httpHandler.NewLocationHandler(locationService, validator.New())

	// Init router
	router, err := httpHandler.NewRouter(&config.Server, l, *pingHandler, *locationHandler)
	if err != nil {
		l.Error("Error initializing router ", zap.Error(err))
		os.Exit(1)
	}

	// Start server
	listenAddr := fmt.Sprintf("%s:%s", config.Server.HttpUrl, config.Server.HttpPort)
	l.Info("Starting the HTTP server", zap.String("listen_address", listenAddr))

	err = http.ListenAndServe(listenAddr, router)
	if err != nil {
		l.Error("Error starting the HTTP server", zap.Error(err))
		os.Exit(1)
	}
}
