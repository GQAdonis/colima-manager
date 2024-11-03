package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gqadonis/colima-manager/internal/config"
	"github.com/gqadonis/colima-manager/internal/domain"
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
	log.Info("Loading configuration...")
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration: %v", err)
	}
	log.Info("Configuration loaded successfully")

	// Initialize repository
	log.Info("Initializing Colima repository...")
	repo, err := colima.NewColimaRepository()
	if err != nil {
		log.Fatal("Failed to initialize repository: %v", err)
	}
	log.Info("Colima repository initialized successfully")

	// Initialize use case
	log.Info("Initializing Colima use case...")
	useCase := usecase.NewColimaUseCase(repo)
	log.Info("Colima use case initialized successfully")

	// If auto flag is set, start the default profile before starting the API server
	if cfg.Server.Auto.Enabled {
		log.Info("Auto flag detected, preparing to start default profile")
		defaultProfile := cfg.Server.Auto.Default
		if defaultProfile == "" {
			defaultProfile = "default"
			log.Info("No default profile specified, using 'default'")
		}

		// Get profile config
		log.Info("Loading configuration for profile: %s", defaultProfile)
		profileCfg, exists := cfg.Profiles[defaultProfile]
		if !exists {
			log.Info("No configuration found for profile '%s', using defaults", defaultProfile)
			profileCfg = config.ProfileConfig{
				CPUs:           4,
				Memory:         8,
				DiskSize:       60,
				VMType:         "vz",
				Runtime:        "containerd",
				NetworkAddress: true,
				Kubernetes:     true,
			}
		}
		log.Info("Profile configuration: CPUs=%d, Memory=%d, DiskSize=%d, VMType=%s, Runtime=%s, NetworkAddress=%v, Kubernetes=%v",
			profileCfg.CPUs, profileCfg.Memory, profileCfg.DiskSize, profileCfg.VMType, profileCfg.Runtime,
			profileCfg.NetworkAddress, profileCfg.Kubernetes)

		// Convert config.ProfileConfig to domain.ColimaConfig
		colimaCfg := domain.ColimaConfig{
			CPUs:           profileCfg.CPUs,
			Memory:         profileCfg.Memory,
			DiskSize:       profileCfg.DiskSize,
			VMType:         profileCfg.VMType,
			Runtime:        profileCfg.Runtime,
			NetworkAddress: profileCfg.NetworkAddress,
			Kubernetes:     profileCfg.Kubernetes,
			Profile:        defaultProfile,
		}

		// Start the profile
		log.Info("Starting Colima profile '%s'...", defaultProfile)
		if err := useCase.Start(context.Background(), colimaCfg); err != nil {
			log.Fatal("Failed to start profile '%s': %v", defaultProfile, err)
		}

		// Wait for profile to be fully ready
		log.Info("Waiting for profile '%s' to be fully ready...", defaultProfile)
		for {
			status, err := useCase.Status(context.Background(), defaultProfile)
			if err != nil {
				log.Error("Error checking profile status: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}
			if status.Status == "Running" {
				log.Info("Profile '%s' is now running with: CPUs=%d, Memory=%d, DiskSize=%d, Kubernetes=%v",
					defaultProfile, status.CPUs, status.Memory, status.DiskSize, status.Kubernetes)
				break
			}
			log.Info("Profile '%s' status: %s, waiting...", defaultProfile, status.Status)
			time.Sleep(2 * time.Second)
		}

		// If Kubernetes is enabled, verify it's ready
		if profileCfg.Kubernetes {
			log.Info("Verifying Kubernetes configuration...")
			_, err := useCase.GetKubeConfig(context.Background(), defaultProfile)
			if err != nil {
				log.Fatal("Failed to verify Kubernetes configuration: %v", err)
			}
			log.Info("Kubernetes configuration verified successfully")
		}

		log.Info("Profile '%s' is fully ready", defaultProfile)
	}

	// Initialize Echo instance
	log.Info("Initializing HTTP server...")
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
	log.Info("PID file created at: %s", pidFile)

	// Start server
	address := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Info("Starting HTTP server at %s", address)
	go func() {
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
	log.Info("Cleaning up...")
	os.Remove(pidFile)
	log.Info("PID file removed")

	// Shutdown
	log.Info("Shutting down HTTP server...")
	if err := e.Shutdown(context.Background()); err != nil {
		log.Fatal("Error during server shutdown: %v", err)
	}
	log.Info("Server shutdown complete")
}
