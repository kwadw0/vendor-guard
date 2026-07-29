package vendors

import (
	"errors"
	"net/http"

	"vendor-guard/middleware"
	"vendor-guard/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type VendorHandler interface {
	CreateVendor(w http.ResponseWriter, r *http.Request)
	GetVendorByID(w http.ResponseWriter, r *http.Request)
	GetAllVendors(w http.ResponseWriter, r *http.Request)
	UpdateVendor(w http.ResponseWriter, r *http.Request)
	DeleteVendor(w http.ResponseWriter, r *http.Request)
}

type vendorHandler struct {
	service   VendorService
	validator *validator.Validate
}

func NewVendorHandler(service VendorService, v *validator.Validate) VendorHandler {
	return &vendorHandler{
		service:   service,
		validator: v,
	}
}

// CreateVendor godoc
//
//	@Summary		Create a new vendor
//	@Description	Creates a new vendor under an organization
//	@Tags			vendors
//	@Accept			json
//	@Produce		json
//	@Param			body	body		CreateVendorDto								true	"Vendor payload"
//	@Success		201		{object}	utils.SuccessResponse{data=vendors.VendorResponse}	"Vendor created successfully"
//	@Failure		400		{object}	utils.ErrorResponse							"Bad request or validation error"
//	@Failure		403		{object}	utils.ErrorResponse							"Access denied"
//	@Failure		500		{object}	utils.ErrorResponse							"Internal server error"
//	@Security		BearerAuth
//	@Router			/api/vendors [post]
func (h *vendorHandler) CreateVendor(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var dto CreateVendorDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "BAD_REQUEST")
		return
	}
	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}
	vendor, err := h.service.CreateVendor(r.Context(), dto, userID)
	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusCreated, "Vendor created successfully", vendor)
}

// GetVendorByID godoc
//
//	@Summary		Get a vendor by ID
//	@Description	Retrieves a single vendor by its UUID
//	@Tags			vendors
//	@Produce		json
//	@Param			id	path		string										true	"Vendor UUID"
//	@Success		200	{object}	utils.SuccessResponse{data=vendors.VendorResponse}	"Vendor retrieved successfully"
//	@Failure		403	{object}	utils.ErrorResponse							"Access denied"
//	@Failure		404	{object}	utils.ErrorResponse							"Vendor not found"
//	@Security		BearerAuth
//	@Router			/api/vendors/{id} [get]
func (h *vendorHandler) GetVendorByID(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")
	vendor, err := h.service.GetVendorByID(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		utils.ErrorJSON(w, http.StatusNotFound, err, "VENDOR_NOT_FOUND")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Vendor retrieved successfully", vendor)
}

// GetAllVendors godoc
//
//	@Summary		List all vendors
//	@Description	Retrieves all vendors scoped to the authenticated user
//	@Tags			vendors
//	@Produce		json
//	@Success		200	{object}	utils.SuccessResponse{data=[]VendorResponse}	"Vendors retrieved successfully"
//	@Failure		403	{object}	utils.ErrorResponse								"Access denied"
//	@Failure		500	{object}	utils.ErrorResponse								"Internal server error"
//	@Security		BearerAuth
//	@Router			/api/vendors [get]
func (h *vendorHandler) GetAllVendors(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	vendors, err := h.service.GetAllVendors(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Vendors retrieved successfully", vendors)
}

// UpdateVendor godoc
//
//	@Summary		Update a vendor
//	@Description	Updates an existing vendor's details by ID
//	@Tags			vendors
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string										true	"Vendor UUID"
//	@Param			body	body		UpdateVendorDto								true	"Updated vendor payload"
//	@Success		200		{object}	utils.SuccessResponse{data=vendors.VendorResponse}	"Vendor updated successfully"
//	@Failure		400		{object}	utils.ErrorResponse							"Bad request or validation error"
//	@Failure		403		{object}	utils.ErrorResponse							"Access denied"
//	@Failure		500		{object}	utils.ErrorResponse							"Internal server error"
//	@Security		BearerAuth
//	@Router			/api/vendors/{id} [put]
func (h *vendorHandler) UpdateVendor(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")
	var dto UpdateVendorDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "BAD_REQUEST")
		return
	}
	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}
	dto.ID = id
	vendor, err := h.service.UpdateVendor(r.Context(), dto, userID)
	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Vendor updated successfully", vendor)
}

// DeleteVendor godoc
//
//	@Summary		Delete a vendor
//	@Description	Permanently deletes a vendor by ID
//	@Tags			vendors
//	@Produce		json
//	@Param			id	path		string										true	"Vendor UUID"
//	@Success		200	{object}	utils.SuccessResponse{data=utils.EmptyData}	"Vendor deleted successfully"
//	@Failure		403	{object}	utils.ErrorResponse							"Access denied"
//	@Failure		500	{object}	utils.ErrorResponse							"Internal server error"
//	@Security		BearerAuth
//	@Router			/api/vendors/{id} [delete]
func (h *vendorHandler) DeleteVendor(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")
	if err := h.service.DeleteVendor(r.Context(), id, userID); err != nil {
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
