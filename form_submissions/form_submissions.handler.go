package form_submissions

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	ListSubmissions(w http.ResponseWriter, r *http.Request)
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
//	@Param			id		path		string										true	"Form UUID"
//	@Param			body	body		CreateFormSubmissionDto						true	"Submission payload"
//	@Success		201		{object}	utils.SuccessResponse{data=FormSubmissionResponse}	"Submission created successfully"
//	@Failure		400		{object}	utils.ErrorResponse							"Bad request or validation error"
//	@Failure		403		{object}	utils.ErrorResponse							"Access denied"
//	@Failure		500		{object}	utils.ErrorResponse							"Internal server error"
//	@Security		BearerAuth
//	@Router			/api/forms/{id}/submissions [post]
func (h *formSubmissionHandler) CreateSubmission(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	formID := chi.URLParam(r, "id")
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
//	@Summary		Get a submission (enriched)
//	@Description	Retrieves a form submission by ID with form/partner enrichment
//	@Tags			form-submissions
//	@Produce		json
//	@Param			id	path		string										true	"Submission UUID"
//	@Success		200	{object}	utils.SuccessResponse{data=EnrichedSubmissionResponse}	"Submission retrieved successfully"
//	@Failure		403	{object}	utils.ErrorResponse							"Access denied"
//	@Failure		404	{object}	utils.ErrorResponse							"Submission not found"
//	@Security		BearerAuth
//	@Router			/api/submissions/{id} [get]
func (h *formSubmissionHandler) GetSubmissionByID(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")
	submission, err := h.service.GetSubmissionEnrichedByID(r.Context(), id, userID)
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
//	@Description	Retrieves all submissions for a form (org view) — legacy, use GET /api/submissions?form_id=
//	@Tags			form-submissions
//	@Produce		json
//	@Param			id	path		string											true	"Form UUID"
//	@Success		200	{object}	utils.SuccessResponse{data=[]FormSubmissionResponse}	"Submissions retrieved successfully"
//	@Failure		403	{object}	utils.ErrorResponse								"Access denied"
//	@Failure		404	{object}	utils.ErrorResponse								"Form not found"
//	@Security		BearerAuth
//	@Router			/api/forms/{id}/submissions [get]
func (h *formSubmissionHandler) GetSubmissionsByFormID(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	formID := chi.URLParam(r, "id")
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

// ListSubmissions godoc
//
//	@Summary		List submissions (org-scoped, enriched, paginated)
//	@Description	Org member = all org submissions; partner = own only. Single enriched query with COUNT(*) OVER().
//	@Tags			form-submissions
//	@Produce		json
//	@Param			status		query		string	false	"Filter by status"	Enums(pending,approved,rejected,revision_required,in_review)
//	@Param			form_id		query		string	false	"Filter by form ID"	Format(uuid)
//	@Param			partner_id	query		string	false	"Filter by partner ID"	Format(uuid)
//	@Param			q			query		string	false	"Search in form title, partner name/email, responses"
//	@Param			page		query		int		false	"Page (1-based, default 1)"	Minimum(1)
//	@Param			limit		query		int		false	"Limit (default 20, max 50)"	Minimum(1)	Maximum(50)
//	@Param			sort		query		string	false	"Sort column"	Enums(submitted_at,created_at)	Default(submitted_at)
//	@Param			order		query		string	false	"Sort order"	Enums(asc,desc)	Default(desc)
//	@Param			date_from	query		string	false	"Submitted after (RFC3339)"	Format(date-time)
//	@Param			date_to		query		string	false	"Submitted before (RFC3339)"	Format(date-time)
//	@Success		200			{object}	utils.SuccessResponse{data=[]EnrichedSubmissionResponse}	"Submissions retrieved successfully"
//	@Failure		401			{object}	utils.ErrorResponse	"Unauthorized"
//	@Security		BearerAuth
//	@Router			/api/submissions [get]
func (h *formSubmissionHandler) ListSubmissions(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	q := r.URL.Query()

	// Parse query with engineering defaults: page 1, limit 20 max 50, sort submitted_at desc
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	status := strings.TrimSpace(q.Get("status"))
	formID := strings.TrimSpace(q.Get("form_id"))
	partnerID := strings.TrimSpace(q.Get("partner_id"))
	search := strings.TrimSpace(q.Get("q"))
	sortCol := strings.TrimSpace(q.Get("sort"))
	order := strings.TrimSpace(strings.ToLower(q.Get("order")))
	dateFrom := strings.TrimSpace(q.Get("date_from"))
	dateTo := strings.TrimSpace(q.Get("date_to"))

	// Validate date ISO8601 if provided
	if dateFrom != "" {
		if _, err := time.Parse(time.RFC3339, dateFrom); err != nil {
			if _, err2 := time.Parse("2006-01-02", dateFrom); err2 != nil {
				utils.ErrorJSON(w, http.StatusBadRequest, errors.New("invalid date_from, use RFC3339"), "VALIDATION_ERROR")
				return
			}
		}
	}
	if dateTo != "" {
		if _, err := time.Parse(time.RFC3339, dateTo); err != nil {
			if _, err2 := time.Parse("2006-01-02", dateTo); err2 != nil {
				utils.ErrorJSON(w, http.StatusBadRequest, errors.New("invalid date_to, use RFC3339"), "VALIDATION_ERROR")
				return
			}
		}
	}

	query := SubmissionListQuery{
		Status:    status,
		FormID:    formID,
		PartnerID: partnerID,
		Q:         search,
		Page:      page,
		Limit:     limit,
		Sort:      sortCol,
		Order:     order,
		DateFrom:  dateFrom,
		DateTo:    dateTo,
	}
	if err := h.validator.Struct(query); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}

	result, err := h.service.ListSubmissions(r.Context(), userID, query)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}

	meta := utils.PaginationMeta{
		Page:       result.Page,
		Limit:      result.PageSize,
		Total:      result.Total,
		TotalPages: (result.Total + result.PageSize - 1) / result.PageSize,
	}
	if result.Total == 0 {
		meta.TotalPages = 0
	}
	utils.WriteJSONWithMeta(w, http.StatusOK, "Submissions retrieved successfully", result.Submissions, meta)
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
