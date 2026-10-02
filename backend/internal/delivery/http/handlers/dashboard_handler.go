package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/skryfon/employee360/backend/internal/delivery/http/response"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	dashtypes "github.com/skryfon/employee360/backend/internal/types/dashboard"
	dashusecase "github.com/skryfon/employee360/backend/internal/usecase/interface/dashboard"
)

// DashboardHandler serves the read-only admin and super-admin dashboards. Role
// guards live in route registration; the tenant comes from the auth context.
type DashboardHandler struct {
	adminUC      dashusecase.AdminDashboardUseCase
	superAdminUC dashusecase.SuperAdminDashboardUseCase
}

// NewDashboardHandler constructs a DashboardHandler.
func NewDashboardHandler(adminUC dashusecase.AdminDashboardUseCase, superAdminUC dashusecase.SuperAdminDashboardUseCase) *DashboardHandler {
	return &DashboardHandler{adminUC: adminUC, superAdminUC: superAdminUC}
}

func writeDashboardError(c *gin.Context, err error) {
	if errors.Is(err, domainerrors.ErrUnauthorized) {
		response.Unauthorized(c, "unauthorized")
		return
	}
	response.Internal(c, "an unexpected error occurred")
}

// Admin returns the caller's tenant dashboard (admin only).
//
// @Summary      Tenant admin dashboard
// @Description  Counts of users, invitations by status, departments and positions for the caller's tenant, plus the 5 most recent invitations.
// @Tags         dashboard
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Envelope{data=dashtypes.AdminDashboardResponse}
// @Failure      401  {object}  response.Envelope
// @Failure      403  {object}  response.Envelope
// @Router       /api/v1/dashboard/admin [get]
func (h *DashboardHandler) Admin(c *gin.Context) {
	tenantID, ok := tenantOnly(c)
	if !ok {
		return
	}
	res, err := h.adminUC.Execute(c.Request.Context(), tenantID)
	if err != nil {
		writeDashboardError(c, err)
		return
	}
	response.Success(c, dashtypes.ToAdminDashboardResponse(res))
}

// SuperAdmin returns the organisation overview (super_admin only).
//
// @Summary      Super admin dashboard
// @Description  Organisation overview of the caller's tenant: tenant info, user counts (total/active/inactive/pending-invited and by role), invitations by status, departments, positions and the 5 most recent invitations. super_admin only; no cross-tenant aggregates.
// @Tags         dashboard
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Envelope{data=dashtypes.SuperAdminDashboardResponse}
// @Failure      401  {object}  response.Envelope
// @Failure      403  {object}  response.Envelope
// @Router       /api/v1/dashboard/super-admin [get]
func (h *DashboardHandler) SuperAdmin(c *gin.Context) {
	tenantID, ok := tenantOnly(c)
	if !ok {
		return
	}
	res, err := h.superAdminUC.Execute(c.Request.Context(), tenantID)
	if err != nil {
		writeDashboardError(c, err)
		return
	}
	response.Success(c, dashtypes.ToSuperAdminDashboardResponse(res))
}
