package form_templates

import (
	"errors"
	"net/http"

	"preuvio/middleware"
	"preuvio/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type FormTemplateHandler interface {
	CreateTemplate(w http.ResponseWriter, r *http.Request)
	GetTemplateByID(w http.ResponseWriter, r *http.Request)
	GetAllTemplates(w http.ResponseWriter, r *http.Request)
	UpdateTemplate(w http.ResponseWriter, r *http.Request)
	DeleteTemplate(w http.ResponseWriter, r *http.Request)
	CloneTemplateToForm(w http.ResponseWriter, r *http.Request)
}

type formTemplateHandler struct {
	service   FormTemplateService
	validator *validator.Validate
}

func NewHandler(service FormTemplateService, v *validator.Validate) FormTemplateHandler {
	return &formTemplateHandler{
		service:   service,
		validator: v,
	}
}

// CreateTemplate godoc
//
//	@Summary		Create a form template
//	@Description	Creates a new global form template
//	@Tags			form-templates
//	@Accept			json
//	@Produce		json
//	@Param			body	body		CreateFormTemplateDto							true	"Template payload"
//	@Success		201		{object}	utils.SuccessResponse{data=FormTemplateResponse}	"Template created successfully"
//	@Failure		400		{object}	utils.ErrorResponse								"Bad request or validation error"
//	@Failure		500		{object}	utils.ErrorResponse								"Internal server error"
//	@Security		BearerAuth
//	@Router			/api/templates [post]
func (h *formTemplateHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var dto CreateFormTemplateDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "BAD_REQUEST")
		return
	}
	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}
	template, err := h.service.CreateTemplate(r.Context(), dto)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusCreated, "Template created successfully", template)
}

// GetTemplateByID godoc
//
//	@Summary		Get a form template
//	@Description	Retrieves a form template by ID
//	@Tags			form-templates
//	@Produce		json
//	@Param			id	path		string											true	"Template UUID"
//	@Success		200	{object}	utils.SuccessResponse{data=FormTemplateResponse}	"Template retrieved successfully"
//	@Failure		404	{object}	utils.ErrorResponse								"Template not found"
//	@Security		BearerAuth
//	@Router			/api/templates/{id} [get]
func (h *formTemplateHandler) GetTemplateByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	template, err := h.service.GetTemplateByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			utils.ErrorJSON(w, http.StatusNotFound, err, "TEMPLATE_NOT_FOUND")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Template retrieved successfully", template)
}

// GetAllTemplates godoc
//
//	@Summary		List all form templates
//	@Description	Retrieves all active form templates
//	@Tags			form-templates
//	@Produce		json
//	@Success		200	{object}	utils.SuccessResponse{data=[]FormTemplateResponse}	"Templates retrieved successfully"
//	@Failure		500	{object}	utils.ErrorResponse									"Internal server error"
//	@Security		BearerAuth
//	@Router			/api/templates [get]
func (h *formTemplateHandler) GetAllTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.service.GetAllTemplates(r.Context())
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Templates retrieved successfully", templates)
}

// UpdateTemplate godoc
//
//	@Summary		Update a form template
//	@Description	Updates an existing form template
//	@Tags			form-templates
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string											true	"Template UUID"
//	@Param			body	body		UpdateFormTemplateDto							true	"Updated template payload"
//	@Success		200		{object}	utils.SuccessResponse{data=FormTemplateResponse}	"Template updated successfully"
//	@Failure		400		{object}	utils.ErrorResponse								"Bad request or validation error"
//	@Failure		404		{object}	utils.ErrorResponse								"Template not found"
//	@Security		BearerAuth
//	@Router			/api/templates/{id} [put]
func (h *formTemplateHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var dto UpdateFormTemplateDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "BAD_REQUEST")
		return
	}
	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}
	template, err := h.service.UpdateTemplate(r.Context(), id, dto)
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			utils.ErrorJSON(w, http.StatusNotFound, err, "TEMPLATE_NOT_FOUND")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Template updated successfully", template)
}

// DeleteTemplate godoc
//
//	@Summary		Delete a form template
//	@Description	Permanently deletes a form template
//	@Tags			form-templates
//	@Produce		json
//	@Param			id	path		string				true	"Template UUID"
//	@Success		200	{object}	utils.SuccessResponse	"Template deleted successfully"
//	@Failure		404	{object}	utils.ErrorResponse		"Template not found"
//	@Security		BearerAuth
//	@Router			/api/templates/{id} [delete]
func (h *formTemplateHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.service.DeleteTemplate(r.Context(), id); err != nil {
		utils.ErrorJSON(w, http.StatusNotFound, err, "TEMPLATE_NOT_FOUND")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Template deleted successfully", nil)
}

// CloneTemplateToForm godoc
//
//	@Summary		Clone a template to a form
//	@Description	Clones a form template into a new org form
//	@Tags			form-templates
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string											true	"Template UUID"
//	@Param			body	body		CloneTemplateDto								true	"Clone payload"
//	@Success		201		{object}	utils.SuccessResponse{data=FormTemplateResponse}	"Form cloned successfully"
//	@Failure		400		{object}	utils.ErrorResponse								"Bad request or validation error"
//	@Failure		404		{object}	utils.ErrorResponse								"Template not found"
//	@Security		BearerAuth
//	@Router			/api/templates/{id}/clone [post]
func (h *formTemplateHandler) CloneTemplateToForm(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")
	var dto CloneTemplateDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "BAD_REQUEST")
		return
	}
	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}
	template, err := h.service.CloneTemplateToForm(r.Context(), id, dto, userID)
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			utils.ErrorJSON(w, http.StatusNotFound, err, "TEMPLATE_NOT_FOUND")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusCreated, "Form cloned successfully", template)
}
