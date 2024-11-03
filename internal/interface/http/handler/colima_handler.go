package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gqadonis/colima-manager/internal/domain"
	"github.com/gqadonis/colima-manager/internal/pkg/logger"
	"github.com/gqadonis/colima-manager/internal/usecase"
	"github.com/labstack/echo/v4"
)

type ColimaHandler struct {
	useCase usecase.ColimaUseCaseInterface
	log     *logger.Logger
}

func NewColimaHandler(useCase usecase.ColimaUseCaseInterface) *ColimaHandler {
	return &ColimaHandler{
		useCase: useCase,
		log:     logger.GetLogger(),
	}
}

func (h *ColimaHandler) CheckDependencies(c echo.Context) error {
	h.log.Info("Handling dependency check request")

	status, err := h.useCase.CheckDependencies(c.Request().Context())
	if err != nil {
		var depErr *domain.DependencyError
		if errors.As(err, &depErr) {
			h.log.Error("Dependency check failed: %v", err)
			return c.JSON(http.StatusServiceUnavailable, map[string]string{
				"error": err.Error(),
			})
		}
		h.log.Error("Internal error during dependency check: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "internal server error",
		})
	}

	h.log.Info("Dependency check completed successfully")
	return c.JSON(http.StatusOK, status)
}

func (h *ColimaHandler) UpdateDependencies(c echo.Context) error {
	h.log.Info("Handling dependency update request")

	if err := h.useCase.UpdateDependencies(c.Request().Context()); err != nil {
		var depErr *domain.DependencyError
		if errors.As(err, &depErr) {
			h.log.Error("Dependency update failed: %v", err)
			return c.JSON(http.StatusServiceUnavailable, map[string]string{
				"error": err.Error(),
			})
		}
		h.log.Error("Internal error during dependency update: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "internal server error",
		})
	}

	h.log.Info("Dependencies updated successfully")
	return c.JSON(http.StatusOK, map[string]string{"status": "dependencies updated"})
}

func (h *ColimaHandler) Status(c echo.Context) error {
	profile := c.QueryParam("profile")
	h.log.Info("Handling status request - Profile: %s", profile)

	status, err := h.useCase.Status(c.Request().Context(), profile)
	if err != nil {
		var (
			profileNotFound    *domain.ProfileNotFoundError
			profileNotStarted  *domain.ProfileNotStartedError
			profileUnreachable *domain.ProfileUnreachableError
			profileMalfunction *domain.ProfileMalfunctionError
			depErr             *domain.DependencyError
		)

		switch {
		case errors.As(err, &profileNotFound):
			h.log.Error("Profile not found: %v", err)
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": err.Error(),
			})
		case errors.As(err, &profileNotStarted):
			h.log.Error("Profile not started: %v", err)
			return c.JSON(http.StatusServiceUnavailable, map[string]string{
				"error": err.Error(),
			})
		case errors.As(err, &profileUnreachable):
			h.log.Error("Profile unreachable: %v", err)
			return c.JSON(http.StatusBadGateway, map[string]string{
				"error": err.Error(),
			})
		case errors.As(err, &profileMalfunction):
			h.log.Error("Profile malfunction: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		case errors.As(err, &depErr):
			h.log.Error("Dependency error: %v", err)
			return c.JSON(http.StatusServiceUnavailable, map[string]string{
				"error": err.Error(),
			})
		default:
			h.log.Error("Internal error during status check: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "internal server error",
			})
		}
	}

	h.log.Info("Status request completed successfully - Profile: %s", profile)
	return c.JSON(http.StatusOK, status)
}

func (h *ColimaHandler) Start(c echo.Context) error {
	h.log.Info("Handling start request")

	var config domain.ColimaConfig
	if err := c.Bind(&config); err != nil {
		h.log.Error("Invalid request body: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	h.log.Debug("Starting Colima with config: %+v", config)

	if err := h.useCase.Start(c.Request().Context(), config); err != nil {
		var (
			profileNotFound *domain.ProfileNotFoundError
			depErr          *domain.DependencyError
		)

		switch {
		case errors.As(err, &profileNotFound):
			h.log.Error("Profile not found: %v", err)
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": err.Error(),
			})
		case errors.As(err, &depErr):
			h.log.Error("Dependency error: %v", err)
			return c.JSON(http.StatusServiceUnavailable, map[string]string{
				"error": err.Error(),
			})
		default:
			h.log.Error("Internal error during start: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}
	}

	h.log.Info("Start request completed successfully - Profile: %s", config.Profile)
	return c.JSON(http.StatusOK, map[string]string{"status": "started"})
}

func (h *ColimaHandler) Stop(c echo.Context) error {
	profile := c.QueryParam("profile")
	h.log.Info("Handling stop request - Profile: %s", profile)

	if err := h.useCase.Stop(c.Request().Context(), profile); err != nil {
		var (
			profileNotFound *domain.ProfileNotFoundError
			depErr          *domain.DependencyError
		)

		switch {
		case errors.As(err, &profileNotFound):
			h.log.Error("Profile not found: %v", err)
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": err.Error(),
			})
		case errors.As(err, &depErr):
			h.log.Error("Dependency error: %v", err)
			return c.JSON(http.StatusServiceUnavailable, map[string]string{
				"error": err.Error(),
			})
		default:
			h.log.Error("Internal error during stop: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}
	}

	h.log.Info("Stop request completed successfully - Profile: %s", profile)
	return c.JSON(http.StatusOK, map[string]string{"status": "stopped"})
}

func (h *ColimaHandler) GetKubeConfig(c echo.Context) error {
	profile := c.QueryParam("profile")
	h.log.Info("Handling kubeconfig request - Profile: %s", profile)

	kubeconfig, err := h.useCase.GetKubeConfig(c.Request().Context(), profile)
	if err != nil {
		var (
			profileNotFound *domain.ProfileNotFoundError
			depErr          *domain.DependencyError
		)

		switch {
		case errors.As(err, &profileNotFound):
			h.log.Error("Profile not found: %v", err)
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": err.Error(),
			})
		case errors.As(err, &depErr):
			h.log.Error("Dependency error: %v", err)
			return c.JSON(http.StatusServiceUnavailable, map[string]string{
				"error": err.Error(),
			})
		default:
			h.log.Error("Internal error during kubeconfig retrieval: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}
	}

	h.log.Info("Kubeconfig request completed successfully - Profile: %s", profile)
	return c.JSON(http.StatusOK, map[string]string{"kubeconfig": kubeconfig})
}

func (h *ColimaHandler) Clean(c echo.Context) error {
	h.log.Info("Handling clean request")

	var req domain.CleanRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("Invalid request body: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	h.log.Debug("Cleaning profile: %s", req.Profile)

	if err := h.useCase.Clean(c.Request().Context(), req); err != nil {
		var (
			profileNotFound *domain.ProfileNotFoundError
			depErr          *domain.DependencyError
		)

		switch {
		case errors.As(err, &profileNotFound):
			h.log.Error("Profile not found: %v", err)
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": err.Error(),
			})
		case errors.As(err, &depErr):
			h.log.Error("Dependency error: %v", err)
			return c.JSON(http.StatusServiceUnavailable, map[string]string{
				"error": err.Error(),
			})
		default:
			h.log.Error("Internal error during clean: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}
	}

	if req.Profile == "" {
		h.log.Info("All profiles cleaned successfully")
		return c.JSON(http.StatusOK, map[string]string{"status": "all profiles cleaned"})
	}

	h.log.Info("Clean request completed successfully - Profile: %s", req.Profile)
	return c.JSON(http.StatusOK, map[string]string{"status": fmt.Sprintf("profile %s cleaned", req.Profile)})
}
