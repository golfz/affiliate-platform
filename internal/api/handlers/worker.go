package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/jonosize/affiliate-platform/internal/logger"
	"github.com/jonosize/affiliate-platform/internal/worker"
)

// WorkerHandler handles worker-related HTTP requests
type WorkerHandler struct {
	worker *worker.PriceRefreshWorker
	logger logger.Logger
}

// NewWorkerHandler creates a new worker handler
func NewWorkerHandler(worker *worker.PriceRefreshWorker, logger logger.Logger) *WorkerHandler {
	return &WorkerHandler{
		worker: worker,
		logger: logger,
	}
}

// TriggerPriceRefresh handles POST /api/admin/worker/refresh-prices
// @Summary Manually trigger price refresh job
// @Description Manually triggers the price refresh worker to update all product prices
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Price refresh triggered successfully"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/admin/worker/refresh-prices [post]
func (h *WorkerHandler) TriggerPriceRefresh(c echo.Context) error {
	if err := h.worker.TriggerManualRefresh(); err != nil {
		h.logger.Error("Failed to trigger price refresh", logger.String("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Internal Server Error",
			"message": "Failed to trigger price refresh",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Price refresh triggered successfully",
	})
}
