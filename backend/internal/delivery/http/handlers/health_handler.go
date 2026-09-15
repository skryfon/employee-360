// Package handlers holds thin Gin request handlers: bind/validate input,
// call a usecase, and write a response via internal/delivery/http/response.
// Handlers must never touch GORM/the database directly.
package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/your-org/your-project/backend/internal/delivery/http/response"
	"github.com/your-org/your-project/backend/shared"
)

// HealthHandler intentionally has no dependencies — it does not check the DB.
type HealthHandler struct{}

// NewHealthHandler constructs a HealthHandler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// healthResponse is the health endpoint's response body.
type healthResponse struct {
	Status string `json:"status"`
	App    string `json:"app"`
}

func (h *HealthHandler) Health(c *gin.Context) {
	response.Success(c, healthResponse{
		Status: "ok",
		App:    shared.AppName,
	})
}
