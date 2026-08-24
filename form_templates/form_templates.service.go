package form_templates

import (
	"context"
	"errors"
	"time"

	"preuvio/internal/repo"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var ErrTemplateNotFound = errors.New("template not found")

type FormTemplateService interface {
	CreateTemplate(ctx context.Context, dto CreateFormTemplateDto) (FormTemplateResponse, error)
	GetTemplateByID(ctx context.Context, id string) (FormTemplateResponse, error)
	GetAllTemplates(ctx context.Context) ([]FormTemplateResponse, error)
	UpdateTemplate(ctx context.Context, id string, dto UpdateFormTemplateDto) (FormTemplateResponse, error)
	DeleteTemplate(ctx context.Context, id string) error
	CloneTemplateToForm(ctx context.Context, templateID string, dto CloneTemplateDto, userID string) (FormTemplateResponse, error)
}

type formTemplateService struct {
	repo *repo.Queries
}

func NewService(queries *repo.Queries) FormTemplateService {
	return &formTemplateService{repo: queries}
}

func (s *formTemplateService) CreateTemplate(ctx context.Context, dto CreateFormTemplateDto) (FormTemplateResponse, error) {
	template, err := s.repo.CreateFormTemplate(ctx, repo.CreateFormTemplateParams{
		Title: dto.Title,
		Description: pgtype.Text{
			String: dto.Description,
			Valid:  dto.Description != "",
		},
		Category: pgtype.Text{
			String: dto.Category,
			Valid:  dto.Category != "",
		},
	})
	if err != nil {
		return FormTemplateResponse{}, err
	}
	return mapTemplateToResponse(template), nil
}

func (s *formTemplateService) GetTemplateByID(ctx context.Context, id string) (FormTemplateResponse, error) {
	templateUUID, err := uuid.Parse(id)
	if err != nil {
		return FormTemplateResponse{}, err
	}
	template, err := s.repo.GetFormTemplateByID(ctx, templateUUID)
	if err != nil {
		return FormTemplateResponse{}, ErrTemplateNotFound
	}
	return mapTemplateToResponse(template), nil
}

func (s *formTemplateService) GetAllTemplates(ctx context.Context) ([]FormTemplateResponse, error) {
	templates, err := s.repo.GetAllFormTemplates(ctx)
	if err != nil {
		return nil, err
	}
	response := make([]FormTemplateResponse, 0, len(templates))
	for _, t := range templates {
		response = append(response, mapTemplateToResponse(t))
	}
	return response, nil
}

func (s *formTemplateService) UpdateTemplate(ctx context.Context, id string, dto UpdateFormTemplateDto) (FormTemplateResponse, error) {
	templateUUID, err := uuid.Parse(id)
	if err != nil {
		return FormTemplateResponse{}, err
	}
	template, err := s.repo.UpdateFormTemplate(ctx, repo.UpdateFormTemplateParams{
		ID: templateUUID,
		Title: dto.Title,
		Description: pgtype.Text{
			String: dto.Description,
			Valid:  dto.Description != "",
		},
		Category: pgtype.Text{
			String: dto.Category,
			Valid:  dto.Category != "",
		},
		IsActive: dto.IsActive,
	})
	if err != nil {
		return FormTemplateResponse{}, err
	}
	return mapTemplateToResponse(template), nil
}

func (s *formTemplateService) DeleteTemplate(ctx context.Context, id string) error {
	templateUUID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return s.repo.DeleteFormTemplate(ctx, templateUUID)
}

func (s *formTemplateService) CloneTemplateToForm(ctx context.Context, templateID string, dto CloneTemplateDto, userID string) (FormTemplateResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return FormTemplateResponse{}, err
	}

	org, err := s.repo.GetOrganizationByUserID(ctx, userUUID)
	if err != nil {
		return FormTemplateResponse{}, ErrTemplateNotFound
	}

	templateUUID, err := uuid.Parse(templateID)
	if err != nil {
		return FormTemplateResponse{}, err
	}

	_, err = s.repo.CloneTemplateToForm(ctx, repo.CloneTemplateToFormParams{
		OrganizationID: org.ID,
		TemplateID: pgtype.UUID{
			Bytes: templateUUID,
			Valid: true,
		},
		Title: dto.Title,
	})
	if err != nil {
		return FormTemplateResponse{}, err
	}
	template, err := s.repo.GetFormTemplateByID(ctx, templateUUID)
	if err != nil {
		return FormTemplateResponse{}, err
	}
	return mapTemplateToResponse(template), nil
}

func mapTemplateToResponse(t repo.FormTemplate) FormTemplateResponse {
	return FormTemplateResponse{
		ID:          t.ID.String(),
		Title:       t.Title,
		Description: t.Description.String,
		Category:    t.Category.String,
		IsActive:    t.IsActive,
		CreatedAt:   t.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   t.UpdatedAt.Time.Format(time.RFC3339),
	}
}
