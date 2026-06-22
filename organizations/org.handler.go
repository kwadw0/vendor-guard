package organizations

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"vendor-guard/middleware"
	"vendor-guard/utils"
)

type OrganizationHandler struct {
	service  OrganizationService
	validate *validator.Validate
}

func NewOrganizationHandler(service OrganizationService, validate *validator.Validate) *OrganizationHandler {
	return &OrganizationHandler{service: service, validate: validate}
}

// CreateOrganization godoc
// @Summary Create a new organization
// @Description Creates a new organization and links it to the authenticated user. Requires Bearer token.
// @Tags organizations
// @Accept json
// @Produce json
// @Param request body CreateOrganizationDto true "Organization data"
// @Success 201 {object} utils.SuccessResponse[OrganizationResponseDto]
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/organizations [post]
func (h *OrganizationHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		utils.ErrorJSON(w, http.StatusUnauthorized, errors.New("unauthorized"), "UNAUTHORIZED")
		return
	}

	var dto CreateOrganizationDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "INVALID_JSON")
		return
	}

	if err := h.validate.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}

	org, err := h.service.CreateOrganization(r.Context(), dto, userID)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, "Organization created successfully", org)
}

// GetOrganizationByUserID godoc
// @Summary Get my organization
// @Description Returns the organization the authenticated user belongs to. Requires Bearer token.
// @Tags organizations
// @Produce json
// @Success 200 {object} utils.SuccessResponse[OrganizationResponseDto]
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/organizations/me [get]
func (h *OrganizationHandler) GetOrganizationByUserID(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		utils.ErrorJSON(w, http.StatusUnauthorized, errors.New("unauthorized"), "UNAUTHORIZED")
		return
	}

	org, err := h.service.GetOrganizationByUserID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrOrganizationNotFound) {
			utils.ErrorJSON(w, http.StatusNotFound, err, "NOT_FOUND")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}

	utils.WriteJSON(w, http.StatusOK, "Organization retrieved successfully", org)
}

// GetOrganizationById godoc
// @Summary Get organization by ID
// @Description Retrieve organization details by its UUID
// @Tags organizations
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} utils.SuccessResponse[OrganizationResponseDto]
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /api/organizations/{id} [get]
func (h *OrganizationHandler) GetOrganizationById(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "INVALID_ID")
		return
	}

	org, err := h.service.GetOrganizationById(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrOrganizationNotFound) {
			utils.ErrorJSON(w, http.StatusNotFound, err, "NOT_FOUND")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}

	utils.WriteJSON(w, http.StatusOK, "Organization retrieved successfully", org)
}

// GetAllOrganizations godoc
// @Summary List all organizations
// @Description Retrieve a list of all organizations
// @Tags organizations
// @Produce json
// @Success 200 {object} utils.SuccessResponse[[]OrganizationResponseDto]
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/organizations [get]
func (h *OrganizationHandler) GetAllOrganizations(w http.ResponseWriter, r *http.Request) {
	orgs, err := h.service.GetAllOrganizations(r.Context())
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Organizations retrieved successfully", orgs)
}

// UpdateOrganization godoc
// @Summary Update an organization
// @Description Update an existing organization by ID
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Param request body UpdateOrganizationDto true "Updated organization data"
// @Success 200 {object} utils.SuccessResponse[OrganizationResponseDto]
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/organizations/{id} [put]
func (h *OrganizationHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "INVALID_ID")
		return
	}

	var dto UpdateOrganizationDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "INVALID_JSON")
		return
	}

	if err := h.validate.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}

	org, err := h.service.UpdateOrganization(r.Context(), id, dto)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}

	utils.WriteJSON(w, http.StatusOK, "Organization updated successfully", org)
}

// DeleteOrganization godoc
// @Summary Delete an organization
// @Description Delete an organization by ID
// @Tags organizations
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} utils.SuccessResponse[utils.EmptyData]
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/organizations/{id} [delete]
func (h *OrganizationHandler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "INVALID_ID")
		return
	}

	if err := h.service.DeleteOrganization(r.Context(), id); err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}

	utils.WriteJSON(w, http.StatusOK, "Organization deleted successfully", utils.EmptyData{})
}
