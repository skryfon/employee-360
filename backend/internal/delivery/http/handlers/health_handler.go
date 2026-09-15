// Package handlers holds HTTP request handlers.
package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/skryfon/employee360/backend/internal/delivery/http/response"
	usecaseinterface "github.com/skryfon/employee360/backend/internal/usecase/interface"
)

// HealthHandler reports application and database health status.
type HealthHandler struct {
	healthUseCase usecaseinterface.HealthUseCase
}

// NewHealthHandler constructs a HealthHandler with the provided usecase.
func NewHealthHandler(healthUseCase usecaseinterface.HealthUseCase) *HealthHandler {
	return &HealthHandler{healthUseCase: healthUseCase}
}

// healthResponse is the health endpoint's response body.
type healthResponse struct {
	Status   string `json:"status"`
	App      string `json:"app"`
	Database string `json:"database"`
}

// Health handles health check requests and returns system status.
//
// @Summary      Report application and database health
// @Description  Returns application and database connectivity status. Used by load balancers, Kubernetes probes, monitoring, and client SDK smoke tests.
// @Tags         health
// @Produce      json
// @Success      200  {object}  response.Envelope
// @Router       /health [get]
// @Router       /healthz [get]
// @Router       /api/v1/health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	result := h.healthUseCase.Execute(c.Request.Context())

	response.Success(c, healthResponse{
		Status:   "ok",
		App:      result.App,
		Database: result.Database,
	})
}
