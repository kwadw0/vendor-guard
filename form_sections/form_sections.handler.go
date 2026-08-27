package form_sections

import (
	"errors"
	"net/http"

	"preuvio/middleware"
	"preuvio/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type SectionHandler interface {
	CreateFormSection(w http.ResponseWriter, r *http.Request)
	CreateTemplateSection(w http.ResponseWriter, r *http.Request)
	GetFormSections(w http.ResponseWriter, r *http.Request)
	GetTemplateSections(w http.ResponseWriter, r *http.Request)
	UpdateSection(w http.ResponseWriter, r *http.Request)
	DeleteSection(w http.ResponseWriter, r *http.Request)
}

type sectionHandler struct {
	service   SectionService
	validator *validator.Validate
}

func NewHandler(service SectionService, v *validator.Validate) SectionHandler {
	return &sectionHandler{service: service, validator: v}
}

// CreateFormSection godoc
//
//	@Summary		Create a form section
//	@Description	Creates a new section within a form
//	@Tags			form-sections
//	@Accept			json
//	@Produce		json
//	@Param			formId	path		string							true	"Form UUID"
//	@Param			body	body		CreateSectionDto				true	"Section payload"
//	@Success		201		{object}	utils.SuccessResponse{data=SectionResponse}	"Section created successfully"
//	@Failure		400		{object}	utils.ErrorResponse				"Bad request or validation error"
//	@Failure		403		{object}	utils.ErrorResponse				"Access denied"
//	@Security		BearerAuth
//	@Router			/api/forms/{formId}/sections [post]
func (h *sectionHandler) CreateFormSection(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	formID := chi.URLParam(r, "formId")
	var dto CreateSectionDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "BAD_REQUEST")
		return
	}
	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}
	section, err := h.service.CreateFormSection(r.Context(), formID, dto, userID)
	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		if errors.Is(err, ErrFormNotFound) {
			utils.ErrorJSON(w, http.StatusNotFound, err, "FORM_NOT_FOUND")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusCreated, "Section created successfully", section)
}

// CreateTemplateSection godoc
//
//	@Summary		Create a template section
//	@Description	Creates a new section within a template
//	@Tags			template-sections
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Template UUID"
//	@Param			body	body		CreateSectionDto				true	"Section payload"
//	@Success		201		{object}	utils.SuccessResponse{data=SectionResponse}	"Section created successfully"
//	@Failure		400		{object}	utils.ErrorResponse				"Bad request or validation error"
//	@Failure		404		{object}	utils.ErrorResponse				"Template not found"
//	@Security		BearerAuth
//	@Router			/api/templates/{id}/sections [post]
func (h *sectionHandler) CreateTemplateSection(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "id")
	var dto CreateSectionDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "BAD_REQUEST")
		return
	}
	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}
	section, err := h.service.CreateTemplateSection(r.Context(), templateID, dto)
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			utils.ErrorJSON(w, http.StatusNotFound, err, "TEMPLATE_NOT_FOUND")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusCreated, "Section created successfully", section)
}

// GetFormSections godoc
//
//	@Summary		List form sections
//	@Description	Retrieves all sections for a form
//	@Tags			form-sections
//	@Produce		json
//	@Param			formId	path		string								true	"Form UUID"
//	@Success		200		{object}	utils.SuccessResponse{data=[]SectionResponse}	"Sections retrieved successfully"
//	@Failure		403		{object}	utils.ErrorResponse					"Access denied"
//	@Security		BearerAuth
//	@Router			/api/forms/{formId}/sections [get]
func (h *sectionHandler) GetFormSections(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	formID := chi.URLParam(r, "formId")
	sections, err := h.service.GetFormSections(r.Context(), formID, userID)
	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Sections retrieved successfully", sections)
}

// GetTemplateSections godoc
//
//	@Summary		List template sections
//	@Description	Retrieves all sections for a template
//	@Tags			template-sections
//	@Produce		json
//	@Param			id		path		string								true	"Template UUID"
//	@Success		200		{object}	utils.SuccessResponse{data=[]SectionResponse}	"Sections retrieved successfully"
//	@Security		BearerAuth
//	@Router			/api/templates/{id}/sections [get]
func (h *sectionHandler) GetTemplateSections(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "id")
	sections, err := h.service.GetTemplateSections(r.Context(), templateID)
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			utils.ErrorJSON(w, http.StatusNotFound, err, "TEMPLATE_NOT_FOUND")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Sections retrieved successfully", sections)
}

// UpdateSection godoc
//
//	@Summary		Update a section
//	@Description	Updates a form or template section
//	@Tags			form-sections
//	@Accept			json
//	@Produce		json
//	@Param			sectionId	path		string							true	"Section UUID"
//	@Param			body		body		UpdateSectionDto				true	"Updated payload"
//	@Success		200			{object}	utils.SuccessResponse{data=SectionResponse}	"Section updated successfully"
//	@Failure		404			{object}	utils.ErrorResponse				"Section not found"
//	@Security		BearerAuth
//	@Router			/api/sections/{sectionId} [put]
func (h *sectionHandler) UpdateSection(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	sectionID := chi.URLParam(r, "sectionId")
	var dto UpdateSectionDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "BAD_REQUEST")
		return
	}
	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}
	section, err := h.service.UpdateSection(r.Context(), sectionID, dto, userID)
	if err != nil {
		if errors.Is(err, ErrSectionNotFound) {
			utils.ErrorJSON(w, http.StatusNotFound, err, "SECTION_NOT_FOUND")
			return
		}
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Section updated successfully", section)
}

// DeleteSection godoc
//
//	@Summary		Delete a section
//	@Description	Permanently deletes a section
//	@Tags			form-sections
//	@Produce		json
//	@Param			sectionId	path		string					true	"Section UUID"
//	@Success		200			{object}	utils.SuccessResponse	"Section deleted successfully"
//	@Failure		404			{object}	utils.ErrorResponse		"Section not found"
//	@Security		BearerAuth
//	@Router			/api/sections/{sectionId} [delete]
func (h *sectionHandler) DeleteSection(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	sectionID := chi.URLParam(r, "sectionId")
	if err := h.service.DeleteSection(r.Context(), sectionID, userID); err != nil {
		if errors.Is(err, ErrSectionNotFound) {
			utils.ErrorJSON(w, http.StatusNotFound, err, "SECTION_NOT_FOUND")
			return
		}
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Section deleted successfully", nil)
}
