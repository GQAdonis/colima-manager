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

	// Create a file to store the PID
	pid := os.Getpid()
	pidFile := "/tmp/colima-manager.pid"
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
		log.Fatal("Failed to write PID file: %v", err)
	}

	// Start server
	go func() {
		address := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := e.Start(address); err != nil {
			log.Info("Shutting down the server")
		}
	}()

	// If in daemon mode, exit the parent process
	if cfg.Server.Daemon {
		log.Info("Started in daemon mode")
		os.Exit(0)
	}

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Cleanup
	os.Remove(pidFile)

	// Shutdown
	if err := e.Shutdown(context.Background()); err != nil {
		log.Fatal("Error during server shutdown: %v", err)
	}
}
