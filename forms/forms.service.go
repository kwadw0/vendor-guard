package forms

import (
	"context"
	"errors"
	"time"

	"preuvio/internal/repo"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var ErrFormNotFound = errors.New("form not found")
var ErrAccessDenied = errors.New("access denied")

type FormService interface {
	CreateForm(ctx context.Context, dto CreateFormDto, userID string) (FormResponse, error)
	GetFormByID(ctx context.Context, id string, userID string) (FormResponse, error)
	GetFormsByOrg(ctx context.Context, userID string) ([]FormResponse, error)
	UpdateForm(ctx context.Context, id string, dto UpdateFormDto, userID string) (FormResponse, error)
	DeleteForm(ctx context.Context, id string, userID string) error
}

type formService struct {
	repo *repo.Queries
}

func NewService(queries *repo.Queries) FormService {
	return &formService{repo: queries}
}

func (s *formService) CreateForm(ctx context.Context, dto CreateFormDto, userID string) (FormResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return FormResponse{}, err
	}

	org, err := s.repo.GetOrganizationByUserID(ctx, userUUID)
	if err != nil {
		return FormResponse{}, ErrAccessDenied
	}

	var templateID pgtype.UUID
	if dto.TemplateID != "" {
		tID, err := uuid.Parse(dto.TemplateID)
		if err != nil {
			return FormResponse{}, err
		}
		templateID = pgtype.UUID{Bytes: tID, Valid: true}
	}

	form, err := s.repo.CreateForm(ctx, repo.CreateFormParams{
		OrganizationID: org.ID,
		TemplateID:     templateID,
		Title:          dto.Title,
		Description: pgtype.Text{
			String: dto.Description,
			Valid:  dto.Description != "",
		},
		Status: "draft",
	})
	if err != nil {
		return FormResponse{}, err
	}
	return mapFormToResponse(form), nil
}

func (s *formService) GetFormByID(ctx context.Context, id string, userID string) (FormResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return FormResponse{}, err
	}

	formUUID, err := uuid.Parse(id)
	if err != nil {
		return FormResponse{}, err
	}

	org, err := s.repo.GetOrganizationByUserID(ctx, userUUID)
	if err != nil {
		return FormResponse{}, ErrAccessDenied
	}

	form, err := s.repo.GetFormByID(ctx, formUUID)
	if err != nil {
		return FormResponse{}, ErrFormNotFound
	}

	if form.OrganizationID != org.ID {
		return FormResponse{}, ErrAccessDenied
	}

	return mapFormToResponse(form), nil
}

func (s *formService) GetFormsByOrg(ctx context.Context, userID string) ([]FormResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	org, err := s.repo.GetOrganizationByUserID(ctx, userUUID)
	if err != nil {
		return nil, ErrAccessDenied
	}

	forms, err := s.repo.GetFormsByOrg(ctx, org.ID)
	if err != nil {
		return nil, err
	}

	response := make([]FormResponse, 0, len(forms))
	for _, f := range forms {
		response = append(response, mapFormToResponse(f))
	}
	return response, nil
}

func (s *formService) UpdateForm(ctx context.Context, id string, dto UpdateFormDto, userID string) (FormResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return FormResponse{}, err
	}

	formUUID, err := uuid.Parse(id)
	if err != nil {
		return FormResponse{}, err
	}

	org, err := s.repo.GetOrganizationByUserID(ctx, userUUID)
	if err != nil {
		return FormResponse{}, ErrAccessDenied
	}

	form, err := s.repo.GetFormByID(ctx, formUUID)
	if err != nil {
		return FormResponse{}, ErrFormNotFound
	}

	if form.OrganizationID != org.ID {
		return FormResponse{}, ErrAccessDenied
	}

	status := form.Status
	if dto.Status != "" {
		status = dto.Status
	}

	updated, err := s.repo.UpdateForm(ctx, repo.UpdateFormParams{
		ID: formUUID,
		Title: dto.Title,
		Description: pgtype.Text{
			String: dto.Description,
			Valid:  dto.Description != "",
		},
		Status: status,
	})
	if err != nil {
		return FormResponse{}, err
	}
	return mapFormToResponse(updated), nil
}

func (s *formService) DeleteForm(ctx context.Context, id string, userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	formUUID, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	org, err := s.repo.GetOrganizationByUserID(ctx, userUUID)
	if err != nil {
		return ErrAccessDenied
	}

	form, err := s.repo.GetFormByID(ctx, formUUID)
	if err != nil {
		return ErrFormNotFound
	}

	if form.OrganizationID != org.ID {
		return ErrAccessDenied
	}

	return s.repo.DeleteForm(ctx, formUUID)
}

func mapFormToResponse(f repo.Form) FormResponse {
	return FormResponse{
		ID:             f.ID.String(),
		OrganizationID: f.OrganizationID.String(),
		TemplateID:     uuid.UUID(f.TemplateID.Bytes).String(),
		Title:          f.Title,
		Description:    f.Description.String,
		Status:         f.Status,
		CreatedAt:      f.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:      f.UpdatedAt.Time.Format(time.RFC3339),
	}
}
