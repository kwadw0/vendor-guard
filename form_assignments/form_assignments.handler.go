package form_assignments

import (
	"errors"
	"net/http"

	"preuvio/middleware"
	"preuvio/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type AssignmentHandler interface {
	AssignForm(w http.ResponseWriter, r *http.Request)
	GetAssignmentsByForm(w http.ResponseWriter, r *http.Request)
	GetAssignmentsByPartner(w http.ResponseWriter, r *http.Request)
	GetMyAssignments(w http.ResponseWriter, r *http.Request)
	RevokeAssignment(w http.ResponseWriter, r *http.Request)
}

type assignmentHandler struct {
	service   AssignmentService
	validator *validator.Validate
}

func NewHandler(service AssignmentService, v *validator.Validate) AssignmentHandler {
	return &assignmentHandler{service: service, validator: v}
}

// AssignForm godoc
//
//	@Summary		Assign a form to partners
//	@Description	Creates assignments for the given partners. Idempotent: already-assigned pairs are reported in skipped. Form must be active, partners must belong to your org and be active.
//	@Tags			form-assignments
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string										true	"Form UUID"
//	@Param			body	body		AssignFormDto								true	"Assignment payload"
//	@Success		201		{object}	utils.SuccessResponse{data=AssignResultResponse}	"Form assigned successfully"
//	@Failure		400		{object}	utils.ErrorResponse							"Bad request or validation error"
//	@Failure		403		{object}	utils.ErrorResponse							"Access denied"
//	@Failure		404		{object}	utils.ErrorResponse							"Form not found"
//	@Failure		409		{object}	utils.ErrorResponse							"Form is not active"
//	@Security		BearerAuth
//	@Router			/api/forms/{id}/assign [post]
func (h *assignmentHandler) AssignForm(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	formID := chi.URLParam(r, "id")
	var dto AssignFormDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "BAD_REQUEST")
		return
	}
	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}
	result, err := h.service.AssignForm(r.Context(), formID, dto, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrAccessDenied):
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
		case errors.Is(err, ErrFormNotFound):
			utils.ErrorJSON(w, http.StatusNotFound, err, "FORM_NOT_FOUND")
		case errors.Is(err, ErrFormNotActive):
			utils.ErrorJSON(w, http.StatusConflict, err, "FORM_NOT_ACTIVE")
		default:
			utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		}
		return
	}
	utils.WriteJSON(w, http.StatusCreated, "Form assigned successfully", result)
}

// GetAssignmentsByForm godoc
//
//	@Summary		List form assignments
//	@Description	Retrieves all assignments for a form with partner enrichment
//	@Tags			form-assignments
//	@Produce		json
//	@Param			id		path		string																	true	"Form UUID"
//	@Param			status	query		string																	false	"Filter by status"	Enums(assigned,submitted,revoked)
//	@Param			page	query		int																		false	"Page (1-based, default 1)"	Minimum(1)
//	@Param			limit	query		int																		false	"Limit (default 20, max 50)"	Minimum(1)	Maximum(50)
//	@Success		200		{object}	utils.SuccessResponse{data=[]AssignmentResponse,meta=utils.PaginationMeta}	"Assignments retrieved successfully"
//	@Failure		403		{object}	utils.ErrorResponse														"Access denied"
//	@Failure		404		{object}	utils.ErrorResponse														"Form not found"
//	@Security		BearerAuth
//	@Router			/api/forms/{id}/assignments [get]
func (h *assignmentHandler) GetAssignmentsByForm(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	formID := chi.URLParam(r, "id")
	status := r.URL.Query().Get("status")
	assignments, err := h.service.GetAssignmentsByForm(r.Context(), formID, status, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrAccessDenied):
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
		case errors.Is(err, ErrFormNotFound):
			utils.ErrorJSON(w, http.StatusNotFound, err, "FORM_NOT_FOUND")
		default:
			utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		}
		return
	}
	page, limit, _ := utils.ParsePagination(r.URL.Query(), utils.DefaultLimit, utils.MaxLimit)
	paged, meta := utils.PaginateSlice(assignments, page, limit)
	utils.WriteJSONWithMeta(w, http.StatusOK, "Assignments retrieved successfully", paged, meta)
}

// GetAssignmentsByPartner godoc
//
//	@Summary		List partner assignments (org view)
//	@Description	Retrieves all assignments for a partner in your organization
//	@Tags			form-assignments
//	@Produce		json
//	@Param			id		path		string																	true	"Partner UUID"
//	@Param			status	query		string																	false	"Filter by status"	Enums(assigned,submitted,revoked)
//	@Param			page	query		int																		false	"Page (1-based, default 1)"	Minimum(1)
//	@Param			limit	query		int																		false	"Limit (default 20, max 50)"	Minimum(1)	Maximum(50)
//	@Success		200		{object}	utils.SuccessResponse{data=[]AssignmentResponse,meta=utils.PaginationMeta}	"Assignments retrieved successfully"
//	@Failure		403		{object}	utils.ErrorResponse														"Access denied"
//	@Security		BearerAuth
//	@Router			/api/partners/{id}/assignments [get]
func (h *assignmentHandler) GetAssignmentsByPartner(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	partnerID := chi.URLParam(r, "id")
	status := r.URL.Query().Get("status")
	assignments, err := h.service.GetAssignmentsByPartner(r.Context(), partnerID, status, userID)
	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	page, limit, _ := utils.ParsePagination(r.URL.Query(), utils.DefaultLimit, utils.MaxLimit)
	paged, meta := utils.PaginateSlice(assignments, page, limit)
	utils.WriteJSONWithMeta(w, http.StatusOK, "Assignments retrieved successfully", paged, meta)
}

// GetMyAssignments godoc
//
//	@Summary		List my assigned forms (vendor work queue)
//	@Description	Retrieves all forms assigned to the authenticated partner user
//	@Tags			form-assignments
//	@Produce		json
//	@Param			status	query		string																	false	"Filter by status"	Enums(assigned,submitted,overdue,revoked)
//	@Param			page	query		int																		false	"Page (1-based, default 1)"	Minimum(1)
//	@Param			limit	query		int																		false	"Limit (default 20, max 50)"	Minimum(1)	Maximum(50)
//	@Success		200		{object}	utils.SuccessResponse{data=[]AssignmentResponse,meta=utils.PaginationMeta}	"Assignments retrieved successfully"
//	@Failure		403		{object}	utils.ErrorResponse														"Access denied"
//	@Security		BearerAuth
//	@Router			/api/assignments/me [get]
func (h *assignmentHandler) GetMyAssignments(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	status := r.URL.Query().Get("status")
	assignments, err := h.service.GetMyAssignments(r.Context(), status, userID)
	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	page, limit, _ := utils.ParsePagination(r.URL.Query(), utils.DefaultLimit, utils.MaxLimit)
	paged, meta := utils.PaginateSlice(assignments, page, limit)
	utils.WriteJSONWithMeta(w, http.StatusOK, "Assignments retrieved successfully", paged, meta)
}

// RevokeAssignment godoc
//
//	@Summary		Revoke a form assignment
//	@Description	Revokes a partner's assignment (history preserved, future submits blocked)
//	@Tags			form-assignments
//	@Produce		json
//	@Param			id			path		string					true	"Form UUID"
//	@Param			partnerId	path		string					true	"Partner UUID"
//	@Success		200			{object}	utils.SuccessResponse	"Assignment revoked successfully"
//	@Failure		403			{object}	utils.ErrorResponse		"Access denied"
//	@Failure		404			{object}	utils.ErrorResponse		"Assignment not found"
//	@Security		BearerAuth
//	@Router			/api/forms/{id}/assignments/{partnerId} [delete]
func (h *assignmentHandler) RevokeAssignment(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	formID := chi.URLParam(r, "id")
	partnerID := chi.URLParam(r, "partnerId")
	if err := h.service.RevokeAssignment(r.Context(), formID, partnerID, userID); err != nil {
		switch {
		case errors.Is(err, ErrAccessDenied):
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
		case errors.Is(err, ErrFormNotFound), errors.Is(err, ErrAssignmentNotFound):
			utils.ErrorJSON(w, http.StatusNotFound, err, "NOT_FOUND")
		default:
			utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		}
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Assignment revoked successfully", nil)
}
