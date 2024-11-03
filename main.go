package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gqadonis/colima-manager/internal/config"
	"github.com/gqadonis/colima-manager/internal/infrastructure/colima"
	"github.com/gqadonis/colima-manager/internal/interface/http/handler"
	"github.com/gqadonis/colima-manager/internal/interface/http/middleware"
	"github.com/gqadonis/colima-manager/internal/pkg/logger"
	"github.com/gqadonis/colima-manager/internal/usecase"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {
	log := logger.GetLogger()
	log.Info("Starting Colima Manager")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration: %v", err)
	}

	// Initialize repository
	repo, err := colima.NewColimaRepository()
	if err != nil {
		log.Fatal("Failed to initialize repository: %v", err)
	}

	// Initialize use case
	useCase := usecase.NewColimaUseCase(repo)

	// Initialize Echo instance
	e := echo.New()

	// Middleware
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(middleware.RequestLogger(log))

	// Initialize handler
	colimaHandler := handler.NewColimaHandler(useCase)

	// Routes
	e.GET("/dependencies", colimaHandler.CheckDependencies)
	e.POST("/dependencies/update", colimaHandler.UpdateDependencies)
	e.GET("/status", colimaHandler.Status)
	e.POST("/start", colimaHandler.Start)
	e.POST("/stop", colimaHandler.Stop)
	e.GET("/kubeconfig", colimaHandler.GetKubeConfig)
	e.POST("/clean", colimaHandler.Clean)

	// Start server
	go func() {
		if err := e.Start(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
			log.Info("Shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Shutdown
	if err := e.Shutdown(context.Background()); err != nil {
		log.Fatal("Error during server shutdown: %v", err)
	}
}
