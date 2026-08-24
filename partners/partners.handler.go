package partners

import (
	"errors"
	"net/http"

	"preuvio/middleware"
	"preuvio/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type PartnerHandler interface {
	CreatePartner(w http.ResponseWriter, r *http.Request)
	GetPartnerByID(w http.ResponseWriter, r *http.Request)
	GetAllPartners(w http.ResponseWriter, r *http.Request)
	UpdatePartner(w http.ResponseWriter, r *http.Request)
	DeletePartner(w http.ResponseWriter, r *http.Request)
}

type partnerHandler struct {
	service   PartnerService
	validator *validator.Validate
}

func NewPartnerHandler(service PartnerService, v *validator.Validate) PartnerHandler {
	return &partnerHandler{
		service:   service,
		validator: v,
	}
}

// CreatePartner godoc
//
//	@Summary		Create a new partner
//	@Description	Creates a new partner under an organization
//	@Tags			partners
//	@Accept			json
//	@Produce		json
//	@Param			body	body		CreatePartnerDto								true	"Partner payload"
//	@Success		201		{object}	utils.SuccessResponse{data=partners.PartnerResponse}	"Partner created successfully"
//	@Failure		400		{object}	utils.ErrorResponse							"Bad request or validation error"
//	@Failure		403		{object}	utils.ErrorResponse							"Access denied"
//	@Failure		500		{object}	utils.ErrorResponse							"Internal server error"
//	@Security		BearerAuth
//	@Router			/api/partners [post]
func (h *partnerHandler) CreatePartner(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var dto CreatePartnerDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "BAD_REQUEST")
		return
	}
	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}
	partner, err := h.service.CreatePartner(r.Context(), dto, userID)
	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusCreated, "Partner created successfully", partner)
}

// GetPartnerByID godoc
//
//	@Summary		Get a partner by ID
//	@Description	Retrieves a single partner by its UUID
//	@Tags			partners
//	@Produce		json
//	@Param			id	path		string										true	"Partner UUID"
//	@Success		200	{object}	utils.SuccessResponse{data=partners.PartnerResponse}	"Partner retrieved successfully"
//	@Failure		403	{object}	utils.ErrorResponse							"Access denied"
//	@Failure		404	{object}	utils.ErrorResponse							"Partner not found"
//	@Security		BearerAuth
//	@Router			/api/partners/{id} [get]
func (h *partnerHandler) GetPartnerByID(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")
	partner, err := h.service.GetPartnerByID(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		utils.ErrorJSON(w, http.StatusNotFound, err, "PARTNER_NOT_FOUND")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Partner retrieved successfully", partner)
}

// GetAllPartners godoc
//
//	@Summary		List all partners
//	@Description	Retrieves all partners scoped to the authenticated user
//	@Tags			partners
//	@Produce		json
//	@Success		200	{object}	utils.SuccessResponse{data=[]PartnerResponse}	"Partners retrieved successfully"
//	@Failure		403	{object}	utils.ErrorResponse								"Access denied"
//	@Failure		500	{object}	utils.ErrorResponse								"Internal server error"
//	@Security		BearerAuth
//	@Router			/api/partners [get]
func (h *partnerHandler) GetAllPartners(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	partners, err := h.service.GetAllPartners(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Partners retrieved successfully", partners)
}

// UpdatePartner godoc
//
//	@Summary		Update a partner
//	@Description	Updates an existing partner's details by ID
//	@Tags			partners
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string										true	"Partner UUID"
//	@Param			body	body		UpdatePartnerDto								true	"Updated partner payload"
//	@Success		200		{object}	utils.SuccessResponse{data=partners.PartnerResponse}	"Partner updated successfully"
//	@Failure		400		{object}	utils.ErrorResponse							"Bad request or validation error"
//	@Failure		403		{object}	utils.ErrorResponse							"Access denied"
//	@Failure		500		{object}	utils.ErrorResponse							"Internal server error"
//	@Security		BearerAuth
//	@Router			/api/partners/{id} [put]
func (h *partnerHandler) UpdatePartner(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")
	var dto UpdatePartnerDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "BAD_REQUEST")
		return
	}
	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}
	dto.ID = id
	partner, err := h.service.UpdatePartner(r.Context(), dto, userID)
	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Partner updated successfully", partner)
}

// DeletePartner godoc
//
//	@Summary		Delete a partner
//	@Description	Permanently deletes a partner by ID
//	@Tags			partners
//	@Produce		json
//	@Param			id	path		string										true	"Partner UUID"
//	@Success		200	{object}	utils.SuccessResponse{data=utils.EmptyData}	"Partner deleted successfully"
//	@Failure		403	{object}	utils.ErrorResponse							"Access denied"
//	@Failure		500	{object}	utils.ErrorResponse							"Internal server error"
//	@Security		BearerAuth
//	@Router			/api/partners/{id} [delete]
func (h *partnerHandler) DeletePartner(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")
	if err := h.service.DeletePartner(r.Context(), id, userID); err != nil {
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
