package handler

import (
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

func (h *ColimaHandler) handleError(c echo.Context, err error) error {
	switch e := err.(type) {
	case *domain.ProfileNotFoundError:
		return c.JSON(http.StatusNotFound, map[string]string{"error": e.Error()})
	case *domain.ProfileNotStartedError:
		return c.JSON(http.StatusBadRequest, map[string]string{"error": e.Error()})
	case *domain.ProfileUnreachableError:
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": e.Error()})
	case *domain.ProfileMalfunctionError:
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": e.Error()})
	case *domain.ProfileBusyError:
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"error": e.Error(),
			"code":  "profile_busy",
		})
	case *domain.DependencyError:
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": e.Error()})
	case *domain.DockerContextError:
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": e.Error()})
	default:
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}
}

func (h *ColimaHandler) CheckDependencies(c echo.Context) error {
	status, err := h.useCase.CheckDependencies(c.Request().Context())
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(http.StatusOK, status)
}

func (h *ColimaHandler) UpdateDependencies(c echo.Context) error {
	if err := h.useCase.UpdateDependencies(c.Request().Context()); err != nil {
		return h.handleError(c, err)
	}
	return c.NoContent(http.StatusOK)
}

func (h *ColimaHandler) Start(c echo.Context) error {
	var config domain.ColimaConfig
	if err := c.Bind(&config); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if err := h.useCase.Start(c.Request().Context(), config); err != nil {
		return h.handleError(c, err)
	}
	return c.NoContent(http.StatusOK)
}

func (h *ColimaHandler) Stop(c echo.Context) error {
	profile := c.QueryParam("profile")
	if err := h.useCase.Stop(c.Request().Context(), profile); err != nil {
		return h.handleError(c, err)
	}
	return c.NoContent(http.StatusOK)
}

func (h *ColimaHandler) Status(c echo.Context) error {
	profile := c.QueryParam("profile")
	status, err := h.useCase.Status(c.Request().Context(), profile)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(http.StatusOK, status)
}

func (h *ColimaHandler) GetKubeConfig(c echo.Context) error {
	profile := c.QueryParam("profile")
	kubeconfig, err := h.useCase.GetKubeConfig(c.Request().Context(), profile)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.String(http.StatusOK, kubeconfig)
}

func (h *ColimaHandler) Clean(c echo.Context) error {
	var req domain.CleanRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if err := h.useCase.Clean(c.Request().Context(), req); err != nil {
		return h.handleError(c, err)
	}
	return c.NoContent(http.StatusOK)
}
