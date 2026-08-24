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
	GetSubmissionsByFormID(ctx context.Context, formID string, userID string) ([]FormSubmissionResponse, error)
	GetSubmissionsByPartnerID(ctx context.Context, partnerID string, userID string) ([]FormSubmissionResponse, error)
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

	// Check if user is org member (reviewer) or partner user (submitter)
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

	// Verify user is the partner user
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
