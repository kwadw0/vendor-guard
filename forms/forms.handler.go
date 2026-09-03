package forms

import (
	"errors"
	"net/http"

	"preuvio/middleware"
	"preuvio/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type FormHandler interface {
	CreateForm(w http.ResponseWriter, r *http.Request)
	GetFormByID(w http.ResponseWriter, r *http.Request)
	GetFormsByOrg(w http.ResponseWriter, r *http.Request)
	UpdateForm(w http.ResponseWriter, r *http.Request)
	DeleteForm(w http.ResponseWriter, r *http.Request)
	GetFormDetail(w http.ResponseWriter, r *http.Request)
}

type formHandler struct {
	service   FormService
	validator *validator.Validate
}

func NewHandler(service FormService, v *validator.Validate) FormHandler {
	return &formHandler{
		service:   service,
		validator: v,
	}
}

// CreateForm godoc
//
//	@Summary		Create a scratch form
//	@Description	Creates a new empty form for the authenticated user's organization. Use POST /templates/{id}/clone to create a form from a template.
//	@Tags			forms
//	@Accept			json
//	@Produce		json
//	@Param			body	body		CreateFormDto					true	"Form payload (scratch build - no template_id)"
//	@Success		201		{object}	utils.SuccessResponse{data=FormResponse}	"Form created successfully"
//	@Failure		400		{object}	utils.ErrorResponse				"Bad request or validation error"
//	@Failure		403		{object}	utils.ErrorResponse				"Access denied"
//	@Failure		500		{object}	utils.ErrorResponse				"Internal server error"
//	@Security		BearerAuth
//	@Router			/api/forms [post]
func (h *formHandler) CreateForm(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var dto CreateFormDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "BAD_REQUEST")
		return
	}
	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}
	form, err := h.service.CreateForm(r.Context(), dto, userID)
	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusCreated, "Form created successfully", form)
}

// GetFormByID godoc
//
//	@Summary		Get a form
//	@Description	Retrieves a form by ID with its sections and fields
//	@Tags			forms
//	@Produce		json
//	@Param			id	path		string							true	"Form UUID"
//	@Success		200	{object}	utils.SuccessResponse{data=FormResponse}	"Form retrieved successfully"
//	@Failure		403	{object}	utils.ErrorResponse				"Access denied"
//	@Failure		404	{object}	utils.ErrorResponse				"Form not found"
//	@Security		BearerAuth
//	@Router			/api/forms/{id} [get]
func (h *formHandler) GetFormByID(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")
	form, err := h.service.GetFormByID(r.Context(), id, userID)
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
	utils.WriteJSON(w, http.StatusOK, "Form retrieved successfully", form)
}

// GetFormsByOrg godoc
//
//	@Summary		List org forms
//	@Description	Retrieves all forms for the authenticated user's organization
//	@Tags			forms
//	@Produce		json
//	@Success		200	{object}	utils.SuccessResponse{data=[]FormResponse}	"Forms retrieved successfully"
//	@Failure		403	{object}	utils.ErrorResponse							"Access denied"
//	@Failure		500	{object}	utils.ErrorResponse							"Internal server error"
//	@Security		BearerAuth
//	@Router			/api/forms [get]
func (h *formHandler) GetFormsByOrg(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	forms, err := h.service.GetFormsByOrg(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Forms retrieved successfully", forms)
}

// UpdateForm godoc
//
//	@Summary		Update a form
//	@Description	Updates an existing form's details
//	@Tags			forms
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Form UUID"
//	@Param			body	body		UpdateFormDto					true	"Updated form payload"
//	@Success		200		{object}	utils.SuccessResponse{data=FormResponse}	"Form updated successfully"
//	@Failure		400		{object}	utils.ErrorResponse				"Bad request or validation error"
//	@Failure		403		{object}	utils.ErrorResponse				"Access denied"
//	@Failure		404		{object}	utils.ErrorResponse				"Form not found"
//	@Security		BearerAuth
//	@Router			/api/forms/{id} [put]
func (h *formHandler) UpdateForm(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")
	var dto UpdateFormDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "BAD_REQUEST")
		return
	}
	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}
	form, err := h.service.UpdateForm(r.Context(), id, dto, userID)
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
	utils.WriteJSON(w, http.StatusOK, "Form updated successfully", form)
}

// DeleteForm godoc
//
//	@Summary		Delete a form
//	@Description	Permanently deletes a form
//	@Tags			forms
//	@Produce		json
//	@Param			id	path		string					true	"Form UUID"
//	@Success		200	{object}	utils.SuccessResponse	"Form deleted successfully"
//	@Failure		403	{object}	utils.ErrorResponse		"Access denied"
//	@Failure		404	{object}	utils.ErrorResponse		"Form not found"
//	@Security		BearerAuth
//	@Router			/api/forms/{id} [delete]
func (h *formHandler) DeleteForm(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")
	if err := h.service.DeleteForm(r.Context(), id, userID); err != nil {
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
	utils.WriteJSON(w, http.StatusOK, "Form deleted successfully", nil)
}

// GetFormDetail godoc
//
//	@Summary		Get full form with sections and fields
//	@Description	Retrieves a form with all its sections and nested fields in one call
//	@Tags			forms
//	@Produce		json
//	@Param			id	path		string								true	"Form UUID"
//	@Success		200	{object}	utils.SuccessResponse{data=FormDetailResponse}	"Form detail retrieved successfully"
//	@Failure		403	{object}	utils.ErrorResponse					"Access denied"
//	@Failure		404	{object}	utils.ErrorResponse					"Form not found"
//	@Security		BearerAuth
//	@Router			/api/forms/{id}/detail [get]
func (h *formHandler) GetFormDetail(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")
	detail, err := h.service.GetFormDetail(r.Context(), id, userID)
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
	utils.WriteJSON(w, http.StatusOK, "Form detail retrieved successfully", detail)
}
