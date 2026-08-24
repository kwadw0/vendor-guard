package form_submissions

import (
	"errors"
	"net/http"

	"preuvio/middleware"
	"preuvio/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type FormSubmissionHandler interface {
	CreateSubmission(w http.ResponseWriter, r *http.Request)
	GetSubmissionByID(w http.ResponseWriter, r *http.Request)
	GetSubmissionsByFormID(w http.ResponseWriter, r *http.Request)
	GetSubmissionsByPartnerID(w http.ResponseWriter, r *http.Request)
	ReviewSubmission(w http.ResponseWriter, r *http.Request)
}

type formSubmissionHandler struct {
	service   FormSubmissionService
	validator *validator.Validate
}

func NewHandler(service FormSubmissionService, v *validator.Validate) FormSubmissionHandler {
	return &formSubmissionHandler{
		service:   service,
		validator: v,
	}
}

// CreateSubmission godoc
//
//	@Summary		Submit a form response
//	@Description	Creates a new form submission from a partner
//	@Tags			form-submissions
//	@Accept			json
//	@Produce		json
//	@Param			formId	path		string										true	"Form UUID"
//	@Param			body	body		CreateFormSubmissionDto						true	"Submission payload"
//	@Success		201		{object}	utils.SuccessResponse{data=FormSubmissionResponse}	"Submission created successfully"
//	@Failure		400		{object}	utils.ErrorResponse							"Bad request or validation error"
//	@Failure		403		{object}	utils.ErrorResponse							"Access denied"
//	@Failure		500		{object}	utils.ErrorResponse							"Internal server error"
//	@Security		BearerAuth
//	@Router			/api/forms/{formId}/submissions [post]
func (h *formSubmissionHandler) CreateSubmission(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	formID := chi.URLParam(r, "formId")
	var dto CreateFormSubmissionDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "BAD_REQUEST")
		return
	}
	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}
	submission, err := h.service.CreateSubmission(r.Context(), formID, dto, userID)
	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusCreated, "Submission created successfully", submission)
}

// GetSubmissionByID godoc
//
//	@Summary		Get a submission
//	@Description	Retrieves a form submission by ID
//	@Tags			form-submissions
//	@Produce		json
//	@Param			id	path		string										true	"Submission UUID"
//	@Success		200	{object}	utils.SuccessResponse{data=FormSubmissionResponse}	"Submission retrieved successfully"
//	@Failure		403	{object}	utils.ErrorResponse							"Access denied"
//	@Failure		404	{object}	utils.ErrorResponse							"Submission not found"
//	@Security		BearerAuth
//	@Router			/api/submissions/{id} [get]
func (h *formSubmissionHandler) GetSubmissionByID(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")
	submission, err := h.service.GetSubmissionByID(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		if errors.Is(err, ErrSubmissionNotFound) {
			utils.ErrorJSON(w, http.StatusNotFound, err, "SUBMISSION_NOT_FOUND")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Submission retrieved successfully", submission)
}

// GetSubmissionsByFormID godoc
//
//	@Summary		List form submissions
//	@Description	Retrieves all submissions for a form (org view)
//	@Tags			form-submissions
//	@Produce		json
//	@Param			formId	path		string											true	"Form UUID"
//	@Success		200		{object}	utils.SuccessResponse{data=[]FormSubmissionResponse}	"Submissions retrieved successfully"
//	@Failure		403		{object}	utils.ErrorResponse								"Access denied"
//	@Failure		404		{object}	utils.ErrorResponse								"Form not found"
//	@Security		BearerAuth
//	@Router			/api/forms/{formId}/submissions [get]
func (h *formSubmissionHandler) GetSubmissionsByFormID(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	formID := chi.URLParam(r, "formId")
	submissions, err := h.service.GetSubmissionsByFormID(r.Context(), formID, userID)
	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Submissions retrieved successfully", submissions)
}

// GetSubmissionsByPartnerID godoc
//
//	@Summary		List partner submissions
//	@Description	Retrieves all submissions for a partner
//	@Tags			form-submissions
//	@Produce		json
//	@Param			partnerId	path		string											true	"Partner UUID"
//	@Success		200			{object}	utils.SuccessResponse{data=[]FormSubmissionResponse}	"Submissions retrieved successfully"
//	@Failure		403			{object}	utils.ErrorResponse								"Access denied"
//	@Security		BearerAuth
//	@Router			/api/partners/{partnerId}/submissions [get]
func (h *formSubmissionHandler) GetSubmissionsByPartnerID(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	partnerID := chi.URLParam(r, "partnerId")
	submissions, err := h.service.GetSubmissionsByPartnerID(r.Context(), partnerID, userID)
	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Submissions retrieved successfully", submissions)
}

// ReviewSubmission godoc
//
//	@Summary		Review a submission
//	@Description	Approves, rejects, or requests revision on a submission
//	@Tags			form-submissions
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string										true	"Submission UUID"
//	@Param			body	body		ReviewFormSubmissionDto						true	"Review payload"
//	@Success		200		{object}	utils.SuccessResponse{data=FormSubmissionResponse}	"Submission reviewed successfully"
//	@Failure		400		{object}	utils.ErrorResponse							"Bad request or validation error"
//	@Failure		403		{object}	utils.ErrorResponse							"Access denied"
//	@Failure		404		{object}	utils.ErrorResponse							"Submission not found"
//	@Security		BearerAuth
//	@Router			/api/submissions/{id}/review [put]
func (h *formSubmissionHandler) ReviewSubmission(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")
	var dto ReviewFormSubmissionDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "BAD_REQUEST")
		return
	}
	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}
	submission, err := h.service.ReviewSubmission(r.Context(), id, dto, userID)
	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			utils.ErrorJSON(w, http.StatusForbidden, err, "FORBIDDEN")
			return
		}
		if errors.Is(err, ErrSubmissionNotFound) {
			utils.ErrorJSON(w, http.StatusNotFound, err, "SUBMISSION_NOT_FOUND")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Submission reviewed successfully", submission)
}
