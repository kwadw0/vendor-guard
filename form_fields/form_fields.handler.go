package form_fields

import (
	"errors"
	"net/http"

	"preuvio/middleware"
	"preuvio/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type FormFieldHandler interface {
	CreateField(w http.ResponseWriter, r *http.Request)
	GetFieldsByFormID(w http.ResponseWriter, r *http.Request)
	UpdateField(w http.ResponseWriter, r *http.Request)
	DeleteField(w http.ResponseWriter, r *http.Request)
}

type formFieldHandler struct {
	service   FormFieldService
	validator *validator.Validate
}

func NewHandler(service FormFieldService, v *validator.Validate) FormFieldHandler {
	return &formFieldHandler{
		service:   service,
		validator: v,
	}
}

// CreateField godoc
//
//	@Summary		Add a field to a form
//	@Description	Creates a new field within a form section
//	@Tags			form-fields
//	@Accept			json
//	@Produce		json
//	@Param			formId	path		string								true	"Form UUID"
//	@Param			body	body		CreateFormFieldDto					true	"Field payload"
//	@Success		201		{object}	utils.SuccessResponse{data=FormFieldResponse}	"Field created successfully"
//	@Failure		400		{object}	utils.ErrorResponse					"Bad request or validation error"
//	@Failure		403		{object}	utils.ErrorResponse					"Access denied"
//	@Failure		500		{object}	utils.ErrorResponse					"Internal server error"
//	@Security		BearerAuth
//	@Router			/api/forms/{formId}/fields [post]
func (h *formFieldHandler) CreateField(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	formID := chi.URLParam(r, "formId")
	var dto CreateFormFieldDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "BAD_REQUEST")
		return
	}
	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}
	field, err := h.service.CreateField(r.Context(), formID, dto, userID)
	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusCreated, "Field created successfully", field)
}

// GetFieldsByFormID godoc
//
//	@Summary		List form fields
//	@Description	Retrieves all fields for a form
//	@Tags			form-fields
//	@Produce		json
//	@Param			formId	path		string									true	"Form UUID"
//	@Success		200		{object}	utils.SuccessResponse{data=[]FormFieldResponse}	"Fields retrieved successfully"
//	@Failure		403		{object}	utils.ErrorResponse						"Access denied"
//	@Failure		404		{object}	utils.ErrorResponse						"Form not found"
//	@Security		BearerAuth
//	@Router			/api/forms/{formId}/fields [get]
func (h *formFieldHandler) GetFieldsByFormID(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	formID := chi.URLParam(r, "formId")
	fields, err := h.service.GetFieldsByFormID(r.Context(), formID, userID)
	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		if errors.Is(err, ErrFieldNotFound) {
			utils.ErrorJSON(w, http.StatusNotFound, err, "FORM_NOT_FOUND")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Fields retrieved successfully", fields)
}

// UpdateField godoc
//
//	@Summary		Update a form field
//	@Description	Updates an existing form field
//	@Tags			form-fields
//	@Accept			json
//	@Produce		json
//	@Param			fieldId	path		string								true	"Field UUID"
//	@Param			body	body		UpdateFormFieldDto					true	"Updated field payload"
//	@Success		200		{object}	utils.SuccessResponse{data=FormFieldResponse}	"Field updated successfully"
//	@Failure		400		{object}	utils.ErrorResponse					"Bad request or validation error"
//	@Failure		403		{object}	utils.ErrorResponse					"Access denied"
//	@Failure		404		{object}	utils.ErrorResponse					"Field not found"
//	@Security		BearerAuth
//	@Router			/api/forms/{formId}/fields/{fieldId} [put]
func (h *formFieldHandler) UpdateField(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	fieldID := chi.URLParam(r, "fieldId")
	var dto UpdateFormFieldDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "BAD_REQUEST")
		return
	}
	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}
	field, err := h.service.UpdateField(r.Context(), fieldID, dto, userID)
	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		if errors.Is(err, ErrFieldNotFound) {
			utils.ErrorJSON(w, http.StatusNotFound, err, "FIELD_NOT_FOUND")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Field updated successfully", field)
}

// DeleteField godoc
//
//	@Summary		Delete a form field
//	@Description	Permanently deletes a form field
//	@Tags			form-fields
//	@Produce		json
//	@Param			fieldId	path		string					true	"Field UUID"
//	@Success		200		{object}	utils.SuccessResponse	"Field deleted successfully"
//	@Failure		403		{object}	utils.ErrorResponse		"Access denied"
//	@Failure		404		{object}	utils.ErrorResponse		"Field not found"
//	@Security		BearerAuth
//	@Router			/api/forms/{formId}/fields/{fieldId} [delete]
func (h *formFieldHandler) DeleteField(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	fieldID := chi.URLParam(r, "fieldId")
	if err := h.service.DeleteField(r.Context(), fieldID, userID); err != nil {
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		if errors.Is(err, ErrFieldNotFound) {
			utils.ErrorJSON(w, http.StatusNotFound, err, "FIELD_NOT_FOUND")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Field deleted successfully", nil)
}
