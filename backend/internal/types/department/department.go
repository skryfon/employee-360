// Package department holds API DTOs for department operations.
package department

import (
	"time"

	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
)

// CreateDepartmentRequest contains fields for creating a department.
// Tenant identity is strictly derived from the request context.
type CreateDepartmentRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
	// IsActive is optional and defaults to true when omitted.
	IsActive *bool `json:"is_active,omitempty"`
}

// UpdateDepartmentRequest contains fields for updating a department.
// Tenant identity is strictly derived from the request context.
type UpdateDepartmentRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
	// IsActive is optional; when omitted the current value is kept.
	IsActive *bool `json:"is_active,omitempty"`
}

// DepartmentResponse represents the department API presentation model.
type DepartmentResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToDepartmentResponse maps a domain Department entity to DepartmentResponse.
func ToDepartmentResponse(d *entity.Department) DepartmentResponse {
	return DepartmentResponse{
		ID:          d.ID,
		Name:        d.Name,
		Description: d.Description,
		IsActive:    d.IsActive,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}
