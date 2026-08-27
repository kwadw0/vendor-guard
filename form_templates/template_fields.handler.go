package form_templates

import (
	"errors"
	"net/http"

	"preuvio/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type TemplateFieldHandler interface {
	CreateField(w http.ResponseWriter, r *http.Request)
	GetFieldsByTemplateID(w http.ResponseWriter, r *http.Request)
	UpdateField(w http.ResponseWriter, r *http.Request)
	DeleteField(w http.ResponseWriter, r *http.Request)
}

type templateFieldHandler struct {
	service   TemplateFieldService
	validator *validator.Validate
}

func NewTemplateFieldHandler(service TemplateFieldService, v *validator.Validate) TemplateFieldHandler {
	return &templateFieldHandler{
		service:   service,
		validator: v,
	}
}

// CreateField godoc
//
//	@Summary		Add a field to a template
//	@Description	Creates a new field within a form template
//	@Tags			template-fields
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string								true	"Template UUID"
//	@Param			body	body		CreateTemplateFieldDto				true	"Field payload"
//	@Success		201		{object}	utils.SuccessResponse{data=TemplateFieldResponse}	"Field created successfully"
//	@Failure		400		{object}	utils.ErrorResponse					"Bad request or validation error"
//	@Failure		404		{object}	utils.ErrorResponse					"Template not found"
//	@Security		BearerAuth
//	@Router			/api/templates/{id}/fields [post]
func (h *templateFieldHandler) CreateField(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "id")
	var dto CreateTemplateFieldDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "BAD_REQUEST")
		return
	}
	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}
	field, err := h.service.CreateField(r.Context(), templateID, dto)
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			utils.ErrorJSON(w, http.StatusNotFound, err, "TEMPLATE_NOT_FOUND")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusCreated, "Field created successfully", field)
}

// GetFieldsByTemplateID godoc
//
//	@Summary		List template fields
//	@Description	Retrieves all fields for a template
//	@Tags			template-fields
//	@Produce		json
//	@Param			id		path		string									true	"Template UUID"
//	@Success		200		{object}	utils.SuccessResponse{data=[]TemplateFieldResponse}	"Fields retrieved successfully"
//	@Failure		404		{object}	utils.ErrorResponse						"Template not found"
//	@Security		BearerAuth
//	@Router			/api/templates/{id}/fields [get]
func (h *templateFieldHandler) GetFieldsByTemplateID(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "id")
	fields, err := h.service.GetFieldsByTemplateID(r.Context(), templateID)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Fields retrieved successfully", fields)
}

// UpdateField godoc
//
//	@Summary		Update a template field
//	@Description	Updates an existing template field
//	@Tags			template-fields
//	@Accept			json
//	@Produce		json
//	@Param			fieldId		path		string								true	"Field UUID"
//	@Param			body		body		UpdateTemplateFieldDto				true	"Updated field payload"
//	@Success		200			{object}	utils.SuccessResponse{data=TemplateFieldResponse}	"Field updated successfully"
//	@Failure		400			{object}	utils.ErrorResponse					"Bad request or validation error"
//	@Failure		404			{object}	utils.ErrorResponse					"Field not found"
//	@Security		BearerAuth
//	@Router			/api/templates/fields/{fieldId} [put]
func (h *templateFieldHandler) UpdateField(w http.ResponseWriter, r *http.Request) {
	fieldID := chi.URLParam(r, "fieldId")
	var dto UpdateTemplateFieldDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "BAD_REQUEST")
		return
	}
	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}
	field, err := h.service.UpdateField(r.Context(), fieldID, dto)
	if err != nil {
		if errors.Is(err, ErrTemplateFieldNotFound) {
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
//	@Summary		Delete a template field
//	@Description	Permanently deletes a template field
//	@Tags			template-fields
//	@Produce		json
//	@Param			fieldId		path		string					true	"Field UUID"
//	@Success		200			{object}	utils.ErrorResponse		"Field deleted successfully"
//	@Failure		404			{object}	utils.ErrorResponse		"Field not found"
//	@Security		BearerAuth
//	@Router			/api/templates/fields/{fieldId} [delete]
func (h *templateFieldHandler) DeleteField(w http.ResponseWriter, r *http.Request) {
	fieldID := chi.URLParam(r, "fieldId")
	if err := h.service.DeleteField(r.Context(), fieldID); err != nil {
		if errors.Is(err, ErrTemplateFieldNotFound) {
			utils.ErrorJSON(w, http.StatusNotFound, err, "FIELD_NOT_FOUND")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Field deleted successfully", nil)
}
