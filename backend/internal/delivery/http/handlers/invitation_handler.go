package handlers

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/delivery/http/response"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	invtypes "github.com/skryfon/employee360/backend/internal/types/invitation"
	invusecase "github.com/skryfon/employee360/backend/internal/usecase/interface/invitation"
)

// InvitationHandler handles HTTP requests for onboarding invitations.
// Tenant and actor identity come exclusively from the request context populated
// by the Auth/Tenant middleware; request bodies never carry a tenant ID.
type InvitationHandler struct {
	inviteUC invusecase.InviteUserUseCase
	acceptUC invusecase.AcceptInvitationUseCase
	resendUC invusecase.ResendInvitationUseCase
	revokeUC invusecase.RevokeInvitationUseCase
	listUC   invusecase.ListInvitationsUseCase
}

// NewInvitationHandler constructs an InvitationHandler.
func NewInvitationHandler(
	inviteUC invusecase.InviteUserUseCase,
	acceptUC invusecase.AcceptInvitationUseCase,
	resendUC invusecase.ResendInvitationUseCase,
	revokeUC invusecase.RevokeInvitationUseCase,
	listUC invusecase.ListInvitationsUseCase,
) *InvitationHandler {
	return &InvitationHandler{inviteUC: inviteUC, acceptUC: acceptUC, resendUC: resendUC, revokeUC: revokeUC, listUC: listUC}
}

// writeInvitationError maps domain errors to HTTP responses.
func writeInvitationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domainerrors.ErrUnauthorized):
		response.Unauthorized(c, "unauthorized")
	case errors.Is(err, domainerrors.ErrForbidden):
		response.Forbidden(c, "insufficient permissions")
	case errors.Is(err, domainerrors.ErrInvitationNotFound):
		response.NotFound(c, "invitation not found")
	case errors.Is(err, domainerrors.ErrInvitationNotPending):
		response.Error(c, 409, "CONFLICT", "invitation is no longer pending")
	case errors.Is(err, domainerrors.ErrEmailAlreadyExists):
		response.Error(c, 409, "CONFLICT", "a user with that email already exists")
	case errors.Is(err, domainerrors.ErrRoleNotFound),
		errors.Is(err, domainerrors.ErrDepartmentNotFound),
		errors.Is(err, domainerrors.ErrPositionNotFound),
		errors.Is(err, domainerrors.ErrInvalidRole),
		errors.Is(err, domainerrors.ErrInvalidEmail):
		response.BadRequest(c, err.Error())
	default:
		response.Internal(c, "an unexpected error occurred")
	}
}

// Invite creates a pending user and invitation (admin only).
//
// @Summary      Invite user
// @Tags         invitations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      invtypes.InviteUserRequest  true  "Invitation details"
// @Success      201      {object}  response.Envelope{data=invtypes.InvitationResponse}
// @Failure      400      {object}  response.Envelope
// @Failure      401      {object}  response.Envelope
// @Failure      403      {object}  response.Envelope
// @Failure      409      {object}  response.Envelope
// @Router       /api/v1/users/invitations [post]
func (h *InvitationHandler) Invite(c *gin.Context) {
	var req invtypes.InviteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request payload")
		return
	}
	inv, err := h.inviteUC.Execute(c.Request.Context(), req)
	if err != nil {
		writeInvitationError(c, err)
		return
	}
	response.Created(c, invtypes.ToInvitationResponse(inv))
}

// Resend reissues the token of a pending invitation (admin only).
//
// @Summary      Resend invitation
// @Tags         invitations
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Invitation ID"
// @Success      200  {object}  response.Envelope{data=invtypes.InvitationResponse}
// @Failure      404  {object}  response.Envelope
// @Failure      409  {object}  response.Envelope
// @Router       /api/v1/users/invitations/{id}/resend [post]
func (h *InvitationHandler) Resend(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid invitation id")
		return
	}
	inv, err := h.resendUC.Execute(c.Request.Context(), id)
	if err != nil {
		writeInvitationError(c, err)
		return
	}
	response.Success(c, invtypes.ToInvitationResponse(inv))
}

// Revoke cancels a pending invitation (admin only).
//
// @Summary      Revoke invitation
// @Tags         invitations
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Invitation ID"
// @Success      200  {object}  response.Envelope
// @Failure      404  {object}  response.Envelope
// @Failure      409  {object}  response.Envelope
// @Router       /api/v1/users/invitations/{id} [delete]
func (h *InvitationHandler) Revoke(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid invitation id")
		return
	}
	if err := h.revokeUC.Execute(c.Request.Context(), id); err != nil {
		writeInvitationError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "invitation revoked"})
}

// List returns the caller's tenant invitations (admin only).
//
// @Summary      List invitations
// @Tags         invitations
// @Produce      json
// @Security     BearerAuth
// @Param        page       query  int  false  "Page (1-based)"
// @Param        page_size  query  int  false  "Page size (max 100)"
// @Success      200  {object}  response.Envelope{data=[]invtypes.InvitationResponse}
// @Router       /api/v1/users/invitations [get]
func (h *InvitationHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	items, total, err := h.listUC.Execute(c.Request.Context(), pageSize, (page-1)*pageSize)
	if err != nil {
		writeInvitationError(c, err)
		return
	}
	out := make([]invtypes.InvitationResponse, 0, len(items))
	for _, i := range items {
		out = append(out, invtypes.ToInvitationResponse(i))
	}
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	response.Paginated(c, out, response.Meta{Page: page, PageSize: pageSize, TotalItems: total, TotalPages: totalPages})
}

// Accept consumes an invitation token and sets the invitee's password (unauthenticated).
//
// @Summary      Accept invitation
// @Tags         invitations
// @Accept       json
// @Produce      json
// @Param        request  body      invtypes.AcceptInvitationRequest  true  "Token and new password"
// @Success      200      {object}  response.Envelope
// @Failure      400      {object}  response.Envelope
// @Router       /api/v1/invitations/accept [post]
func (h *InvitationHandler) Accept(c *gin.Context) {
	var req invtypes.AcceptInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request payload")
		return
	}
	if err := h.acceptUC.Execute(c.Request.Context(), req); err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrInvalidToken):
			response.BadRequest(c, "invalid or expired invitation token")
		case errors.Is(err, domainerrors.ErrInvalidPassword):
			response.BadRequest(c, "password must be at least 8 characters")
		default:
			response.Internal(c, "an unexpected error occurred")
		}
		return
	}
	response.Success(c, gin.H{"message": "invitation accepted"})
}
