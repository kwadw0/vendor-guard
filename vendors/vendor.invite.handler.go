package vendors

import (
	"errors"
	"net/http"

	"vendor-guard/middleware"
	"vendor-guard/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type InviteHandler interface {
	InviteVendorUser(w http.ResponseWriter, r *http.Request)
	AcceptInvitation(w http.ResponseWriter, r *http.Request)
	GetVendorInvitations(w http.ResponseWriter, r *http.Request)
}

type inviteHandler struct {
	service   InviteService
	validator *validator.Validate
}

func NewInviteHandler(s InviteService, v *validator.Validate) InviteHandler {
	return &inviteHandler{service: s, validator: v}
}

// InviteVendorUser godoc
//
//	@Summary		Invite a vendor user
//	@Description	Creates an invitation for a vendor user to join the platform. Requires organization role.
//	@Tags			vendors
//	@Accept			json
//	@Produce		json
//	@Param			vendorId	path	string					true	"Vendor UUID"
//	@Param			body		body	InviteVendorUserDto		true	"Invitation payload"
//	@Success		201			{object}	utils.SuccessResponse{data=vendors.InvitationResponse}	"Invitation created"
//	@Failure		400			{object}	utils.ErrorResponse	"Bad request or validation error"
//	@Failure		403			{object}	utils.ErrorResponse	"Access denied"
//	@Failure		500			{object}	utils.ErrorResponse	"Internal server error"
//	@Security		BearerAuth
//	@Router			/api/vendors/{vendorId}/invite [post]
func (h *inviteHandler) InviteVendorUser(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	vendorID := chi.URLParam(r, "vendorId")
	if vendorID == "" {
		utils.ErrorJSON(w, http.StatusBadRequest, errors.New("vendor id is required"), "BAD_REQUEST")
		return
	}

	var dto InviteVendorUserDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "BAD_REQUEST")
		return
	}
	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}

	invitation, err := h.service.InviteUser(r.Context(), vendorID, userID, dto)
	if err != nil {
		if errors.Is(err, ErrVendorMismatch) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, "Invitation created successfully", invitation)
}

// AcceptInvitation godoc
//
//	@Summary		Accept a vendor invitation
//	@Description	Accepts an invitation to join a vendor. Creates account and returns JWT tokens. Public endpoint.
//	@Tags			vendors
//	@Accept			json
//	@Produce		json
//	@Param			body	body		AcceptInviteDto								true	"Accept invitation payload"
//	@Success		200		{object}	utils.SuccessResponse{data=vendors.TokenResponse}	"Invitation accepted, tokens returned"
//	@Failure		400		{object}	utils.ErrorResponse	"Bad request or validation error"
//	@Failure		404		{object}	utils.ErrorResponse	"Invitation not found or expired"
//	@Failure		409		{object}	utils.ErrorResponse	"User already exists with an organization"
//	@Failure		500		{object}	utils.ErrorResponse	"Internal server error"
//	@Router			/api/vendors/invite/accept [post]
func (h *inviteHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	var dto AcceptInviteDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "BAD_REQUEST")
		return
	}
	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}

	tokens, err := h.service.AcceptInvitation(r.Context(), dto)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvitationNotFound), errors.Is(err, ErrInvitationExpired):
			utils.ErrorJSON(w, http.StatusNotFound, err, "INVITATION_INVALID")
		case errors.Is(err, ErrInvitationAlreadyUsed):
			utils.ErrorJSON(w, http.StatusConflict, err, "INVITATION_USED")
		case errors.Is(err, ErrVendorUserExists):
			utils.ErrorJSON(w, http.StatusConflict, err, "USER_EXISTS")
		default:
			utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		}
		return
	}

	utils.WriteJSON(w, http.StatusOK, "Invitation accepted successfully", tokens)
}

// GetVendorInvitations godoc
//
//	@Summary		List invitations for a vendor
//	@Description	Retrieves all invitations sent for a specific vendor.
//	@Tags			vendors
//	@Produce		json
//	@Param			vendorId	path	string	true	"Vendor UUID"
//	@Success		200	{object}	utils.SuccessResponse{data=[]InvitationResponse}	"Invitations retrieved"
//	@Failure		500	{object}	utils.ErrorResponse	"Internal server error"
//	@Security		BearerAuth
//	@Router			/api/vendors/{vendorId}/invite [get]
func (h *inviteHandler) GetVendorInvitations(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	vendorID := chi.URLParam(r, "vendorId")
	if vendorID == "" {
		utils.ErrorJSON(w, http.StatusBadRequest, errors.New("vendor id is required"), "BAD_REQUEST")
		return
	}

	invitations, err := h.service.GetInvitationsByVendor(r.Context(), vendorID, userID)
	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}

	utils.WriteJSON(w, http.StatusOK, "Invitations retrieved successfully", invitations)
}
