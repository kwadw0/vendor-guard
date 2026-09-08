package form_submissions

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"preuvio/internal/repo"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var ErrSubmissionNotFound = errors.New("submission not found")
var ErrAccessDenied = errors.New("access denied")

type FormSubmissionService interface {
	CreateSubmission(ctx context.Context, formID string, dto CreateFormSubmissionDto, userID string) (FormSubmissionResponse, error)
	GetSubmissionByID(ctx context.Context, id string, userID string) (FormSubmissionResponse, error)
	GetSubmissionEnrichedByID(ctx context.Context, id string, userID string) (EnrichedSubmissionResponse, error)
	GetSubmissionsByFormID(ctx context.Context, formID string, userID string) ([]FormSubmissionResponse, error)
	GetSubmissionsByPartnerID(ctx context.Context, partnerID string, userID string) ([]FormSubmissionResponse, error)
	ListSubmissions(ctx context.Context, userID string, q SubmissionListQuery) (PaginatedSubmissionsResponse, error)
	ReviewSubmission(ctx context.Context, id string, dto ReviewFormSubmissionDto, userID string) (FormSubmissionResponse, error)
}

type formSubmissionService struct {
	repo *repo.Queries
}

func NewService(queries *repo.Queries) FormSubmissionService {
	return &formSubmissionService{repo: queries}
}

func (s *formSubmissionService) CreateSubmission(ctx context.Context, formID string, dto CreateFormSubmissionDto, userID string) (FormSubmissionResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return FormSubmissionResponse{}, err
	}

	formUUID, err := uuid.Parse(formID)
	if err != nil {
		return FormSubmissionResponse{}, err
	}

	partner, err := s.repo.GetPartnerByUserID(ctx, userUUID)
	if err != nil {
		return FormSubmissionResponse{}, ErrAccessDenied
	}

	responsesJSON, err := json.Marshal(dto.Responses)
	if err != nil {
		return FormSubmissionResponse{}, err
	}

	now := time.Now()
	submission, err := s.repo.CreateFormSubmission(ctx, repo.CreateFormSubmissionParams{
		FormID:    formUUID,
		PartnerID: partner.ID,
		SubmittedBy: pgtype.UUID{
			Bytes: userUUID,
			Valid: true,
		},
		Status: "pending",
		Responses: responsesJSON,
		SubmittedAt: pgtype.Timestamptz{
			Time:  now,
			Valid: true,
		},
	})
	if err != nil {
		return FormSubmissionResponse{}, err
	}
	return mapSubmissionToResponse(submission), nil
}

func (s *formSubmissionService) GetSubmissionByID(ctx context.Context, id string, userID string) (FormSubmissionResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return FormSubmissionResponse{}, err
	}

	submissionUUID, err := uuid.Parse(id)
	if err != nil {
		return FormSubmissionResponse{}, err
	}

	submission, err := s.repo.GetFormSubmissionByID(ctx, submissionUUID)
	if err != nil {
		return FormSubmissionResponse{}, ErrSubmissionNotFound
	}

	org, orgErr := s.repo.GetOrganizationByUserID(ctx, userUUID)
	if orgErr == nil {
		form, err := s.repo.GetFormByID(ctx, submission.FormID)
		if err != nil || form.OrganizationID != org.ID {
			return FormSubmissionResponse{}, ErrAccessDenied
		}
		return mapSubmissionToResponse(submission), nil
	}

	partner, pErr := s.repo.GetPartnerByUserID(ctx, userUUID)
	if pErr == nil && submission.PartnerID == partner.ID {
		return mapSubmissionToResponse(submission), nil
	}

	return FormSubmissionResponse{}, ErrAccessDenied
}

func (s *formSubmissionService) GetSubmissionEnrichedByID(ctx context.Context, id string, userID string) (EnrichedSubmissionResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return EnrichedSubmissionResponse{}, err
	}
	submissionUUID, err := uuid.Parse(id)
	if err != nil {
		return EnrichedSubmissionResponse{}, err
	}
	// Verify access via same logic as GetSubmissionByID but use enriched row
	row, err := s.repo.GetSubmissionEnrichedByID(ctx, submissionUUID)
	if err != nil {
		return EnrichedSubmissionResponse{}, ErrSubmissionNotFound
	}
	// Access check
	org, orgErr := s.repo.GetOrganizationByUserID(ctx, userUUID)
	if orgErr == nil {
		// org member: verify submission belongs to org via form org
		form, err := s.repo.GetFormByID(ctx, row.FormID)
		if err != nil || form.OrganizationID != org.ID {
			return EnrichedSubmissionResponse{}, ErrAccessDenied
		}
		return mapEnrichedRowToResponse(row), nil
	}
	partner, pErr := s.repo.GetPartnerByUserID(ctx, userUUID)
	if pErr == nil && row.PartnerID == partner.ID {
		return mapEnrichedRowToResponse(row), nil
	}
	return EnrichedSubmissionResponse{}, ErrAccessDenied
}

func (s *formSubmissionService) GetSubmissionsByFormID(ctx context.Context, formID string, userID string) ([]FormSubmissionResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	formUUID, err := uuid.Parse(formID)
	if err != nil {
		return nil, err
	}

	org, err := s.repo.GetOrganizationByUserID(ctx, userUUID)
	if err != nil {
		return nil, ErrAccessDenied
	}

	form, err := s.repo.GetFormByID(ctx, formUUID)
	if err != nil || form.OrganizationID != org.ID {
		return nil, ErrAccessDenied
	}

	submissions, err := s.repo.GetFormSubmissionsByFormID(ctx, formUUID)
	if err != nil {
		return nil, err
	}

	response := make([]FormSubmissionResponse, 0, len(submissions))
	for _, sub := range submissions {
		response = append(response, mapSubmissionToResponse(sub))
	}
	return response, nil
}

func (s *formSubmissionService) GetSubmissionsByPartnerID(ctx context.Context, partnerID string, userID string) ([]FormSubmissionResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	partnerUUID, err := uuid.Parse(partnerID)
	if err != nil {
		return nil, err
	}

	partner, err := s.repo.GetPartnerByUserID(ctx, userUUID)
	if err != nil || partner.ID != partnerUUID {
		return nil, ErrAccessDenied
	}

	submissions, err := s.repo.GetFormSubmissionsByPartnerID(ctx, partnerUUID)
	if err != nil {
		return nil, err
	}

	response := make([]FormSubmissionResponse, 0, len(submissions))
	for _, sub := range submissions {
		response = append(response, mapSubmissionToResponse(sub))
	}
	return response, nil
}

func (s *formSubmissionService) ListSubmissions(ctx context.Context, userID string, q SubmissionListQuery) (PaginatedSubmissionsResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return PaginatedSubmissionsResponse{}, err
	}

	// Resolve scoping: org member gets all org submissions, partner gets own only
	var orgID uuid.UUID
	var forcedPartnerID *uuid.UUID

	org, orgErr := s.repo.GetOrganizationByUserID(ctx, userUUID)
	if orgErr == nil {
		orgID = org.ID
	} else {
		partner, pErr := s.repo.GetPartnerByUserID(ctx, userUUID)
		if pErr == nil {
			// Partner role: scope to their organization and force partner_id
			orgID = partner.OrganizationID
			forcedPartnerID = &partner.ID
		} else {
			// No org and no partner: return 200 empty per spec (user decided 200 empty not 403)
			return PaginatedSubmissionsResponse{
				Submissions: []EnrichedSubmissionResponse{},
				Total:       0,
				Page:        1,
				PageSize:    20,
			}, nil
		}
	}

	// Defaults: page 1, limit 20, max 50, sort submitted_at desc
	page := q.Page
	if page < 1 {
		page = 1
	}
	limit := q.Limit
	if limit < 1 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	sortCol := q.Sort
	if sortCol == "" {
		sortCol = "submitted_at"
	}
	sortOrder := q.Order
	if sortOrder == "" {
		sortOrder = "desc"
	}
	offset := int32((page - 1) * limit)

	// Build nullable pgtype params
	var formID pgtype.UUID
	if forcedPartnerID != nil {
		// Partner role: ignore q.PartnerID, force own
		formID = parseNullableUUID(q.FormID)
		// partner param will be forced below
	} else {
		formID = parseNullableUUID(q.FormID)
	}
	partnerID := parseNullableUUID(q.PartnerID)
	if forcedPartnerID != nil {
		partnerID = pgtype.UUID{Bytes: *forcedPartnerID, Valid: true}
	}
	status := pgtype.Text{String: q.Status, Valid: q.Status != ""}
	queryText := pgtype.Text{String: q.Q, Valid: q.Q != ""}
	dateFrom := parseNullableTimestamptz(q.DateFrom)
	dateTo := parseNullableTimestamptz(q.DateTo)

	rows, err := s.repo.ListSubmissionsEnriched(ctx, repo.ListSubmissionsEnrichedParams{
		OrganizationID: orgID,
		FormID:         formID,
		Status:         status,
		PartnerID:      partnerID,
		Q:              queryText,
		DateFrom:       dateFrom,
		DateTo:         dateTo,
		SortCol:        sortCol,
		SortOrder:      sortOrder,
		LimitVal:       int32(limit),
		OffsetVal:      offset,
	})
	if err != nil {
		return PaginatedSubmissionsResponse{}, err
	}

	total := 0
	if len(rows) > 0 {
		total = int(rows[0].TotalCount)
	}
	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
		_ = totalPages
	}

	enriched := make([]EnrichedSubmissionResponse, 0, len(rows))
	for _, r := range rows {
		enriched = append(enriched, EnrichedSubmissionResponse{
			FormSubmissionResponse: mapSubmissionToResponse(repo.FormSubmission{
				ID:          r.ID,
				FormID:      r.FormID,
				PartnerID:   r.PartnerID,
				SubmittedBy: r.SubmittedBy,
				Status:      r.Status,
				Responses:   r.Responses,
				SubmittedAt: r.SubmittedAt,
				ReviewedAt:  r.ReviewedAt,
				ReviewedBy:  r.ReviewedBy,
				CreatedAt:   r.CreatedAt,
				UpdatedAt:   r.UpdatedAt,
			}),
			FormTitle:    r.FormTitle,
			FormStatus:   r.FormStatus,
			PartnerName:  r.PartnerName,
			PartnerEmail: r.PartnerEmail,
		})
	}

	return PaginatedSubmissionsResponse{
		Submissions: enriched,
		Total:       total,
		Page:        page,
		PageSize:    limit,
	}, nil
}

func (s *formSubmissionService) ReviewSubmission(ctx context.Context, id string, dto ReviewFormSubmissionDto, userID string) (FormSubmissionResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return FormSubmissionResponse{}, err
	}

	submissionUUID, err := uuid.Parse(id)
	if err != nil {
		return FormSubmissionResponse{}, err
	}

	org, err := s.repo.GetOrganizationByUserID(ctx, userUUID)
	if err != nil {
		return FormSubmissionResponse{}, ErrAccessDenied
	}

	submission, err := s.repo.GetFormSubmissionByID(ctx, submissionUUID)
	if err != nil {
		return FormSubmissionResponse{}, ErrSubmissionNotFound
	}

	form, err := s.repo.GetFormByID(ctx, submission.FormID)
	if err != nil || form.OrganizationID != org.ID {
		return FormSubmissionResponse{}, ErrAccessDenied
	}

	updated, err := s.repo.ReviewFormSubmission(ctx, repo.ReviewFormSubmissionParams{
		ID: submissionUUID,
		Status: dto.Status,
		ReviewedBy: pgtype.UUID{
			Bytes: userUUID,
			Valid: true,
		},
	})
	if err != nil {
		return FormSubmissionResponse{}, err
	}
	return mapSubmissionToResponse(updated), nil
}

func parseNullableUUID(s string) pgtype.UUID {
	if s == "" {
		return pgtype.UUID{Valid: false}
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: id, Valid: true}
}

func parseNullableTimestamptz(s string) pgtype.Timestamptz {
	if s == "" {
		return pgtype.Timestamptz{Valid: false}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		// Try date-only
		t2, err2 := time.Parse("2006-01-02", s)
		if err2 != nil {
			return pgtype.Timestamptz{Valid: false}
		}
		t = t2
	}
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func mapEnrichedRowToResponse(r repo.GetSubmissionEnrichedByIDRow) EnrichedSubmissionResponse {
	var responses map[string]interface{}
	if r.Responses != nil {
		_ = json.Unmarshal(r.Responses, &responses)
	}
	return EnrichedSubmissionResponse{
		FormSubmissionResponse: FormSubmissionResponse{
			ID:          r.ID.String(),
			FormID:      r.FormID.String(),
			PartnerID:   r.PartnerID.String(),
			SubmittedBy: uuid.UUID(r.SubmittedBy.Bytes).String(),
			Status:      r.Status,
			Responses:   responses,
			SubmittedAt: formatOptionalTime(r.SubmittedAt.Time),
			ReviewedAt:  formatOptionalTime(r.ReviewedAt.Time),
			ReviewedBy:  uuid.UUID(r.ReviewedBy.Bytes).String(),
			CreatedAt:   r.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt:   r.UpdatedAt.Time.Format(time.RFC3339),
		},
		FormTitle:    r.FormTitle,
		FormStatus:   r.FormStatus,
		PartnerName:  r.PartnerName,
		PartnerEmail: r.PartnerEmail,
	}
}

func mapSubmissionToResponse(s repo.FormSubmission) FormSubmissionResponse {
	var responses map[string]interface{}
	if s.Responses != nil {
		_ = json.Unmarshal(s.Responses, &responses)
	}

	return FormSubmissionResponse{
		ID:          s.ID.String(),
		FormID:      s.FormID.String(),
		PartnerID:   s.PartnerID.String(),
		SubmittedBy: uuid.UUID(s.SubmittedBy.Bytes).String(),
		Status:      s.Status,
		Responses:   responses,
		SubmittedAt: formatOptionalTime(s.SubmittedAt.Time),
		ReviewedAt:  formatOptionalTime(s.ReviewedAt.Time),
		ReviewedBy:  uuid.UUID(s.ReviewedBy.Bytes).String(),
		CreatedAt:   s.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   s.UpdatedAt.Time.Format(time.RFC3339),
	}
}
