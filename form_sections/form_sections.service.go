package form_sections

import (
	"context"
	"errors"
	"time"

	"preuvio/internal/repo"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var ErrSectionNotFound = errors.New("section not found")
var ErrAccessDenied = errors.New("access denied")
var ErrTemplateNotFound = errors.New("template not found")
var ErrFormNotFound = errors.New("form not found")

type SectionService interface {
	CreateFormSection(ctx context.Context, formID string, dto CreateSectionDto, userID string) (SectionResponse, error)
	CreateTemplateSection(ctx context.Context, templateID string, dto CreateSectionDto) (SectionResponse, error)
	GetFormSections(ctx context.Context, formID string, userID string) ([]SectionResponse, error)
	GetTemplateSections(ctx context.Context, templateID string) ([]SectionResponse, error)
	UpdateSection(ctx context.Context, sectionID string, dto UpdateSectionDto, userID string) (SectionResponse, error)
	DeleteSection(ctx context.Context, sectionID string, userID string) error
}

type sectionService struct {
	repo *repo.Queries
}

func NewService(queries *repo.Queries) SectionService {
	return &sectionService{repo: queries}
}

func (s *sectionService) CreateFormSection(ctx context.Context, formID string, dto CreateSectionDto, userID string) (SectionResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return SectionResponse{}, err
	}
	formUUID, err := uuid.Parse(formID)
	if err != nil {
		return SectionResponse{}, err
	}
	org, err := s.repo.GetOrganizationByUserID(ctx, userUUID)
	if err != nil {
		return SectionResponse{}, ErrAccessDenied
	}
	form, err := s.repo.GetFormByID(ctx, formUUID)
	if err != nil {
		return SectionResponse{}, ErrFormNotFound
	}
	if form.OrganizationID != org.ID {
		return SectionResponse{}, ErrAccessDenied
	}
	section, err := s.repo.CreateFormSection(ctx, repo.CreateFormSectionParams{
		FormID:     pgtype.UUID{Bytes: formUUID, Valid: true},
		TemplateID: pgtype.UUID{Valid: false},
		Title:      dto.Title,
		Description: pgtype.Text{
			String: dto.Description,
			Valid:  dto.Description != "",
		},
		SortOrder: int32(dto.SortOrder),
	})
	if err != nil {
		return SectionResponse{}, err
	}
	return mapSectionToResponse(section), nil
}

func (s *sectionService) CreateTemplateSection(ctx context.Context, templateID string, dto CreateSectionDto) (SectionResponse, error) {
	templateUUID, err := uuid.Parse(templateID)
	if err != nil {
		return SectionResponse{}, err
	}
	if _, err := s.repo.GetFormTemplateByID(ctx, templateUUID); err != nil {
		return SectionResponse{}, ErrTemplateNotFound
	}
	section, err := s.repo.CreateFormSection(ctx, repo.CreateFormSectionParams{
		FormID:     pgtype.UUID{Valid: false},
		TemplateID: pgtype.UUID{Bytes: templateUUID, Valid: true},
		Title:      dto.Title,
		Description: pgtype.Text{
			String: dto.Description,
			Valid:  dto.Description != "",
		},
		SortOrder: int32(dto.SortOrder),
	})
	if err != nil {
		return SectionResponse{}, err
	}
	return mapSectionToResponse(section), nil
}

func (s *sectionService) GetFormSections(ctx context.Context, formID string, userID string) ([]SectionResponse, error) {
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
	if err != nil {
		return nil, ErrFormNotFound
	}
	if form.OrganizationID != org.ID {
		return nil, ErrAccessDenied
	}
	sections, err := s.repo.GetFormSectionsByFormID(ctx, pgtype.UUID{Bytes: formUUID, Valid: true})
	if err != nil {
		return nil, err
	}
	resp := make([]SectionResponse, 0, len(sections))
	for _, sec := range sections {
		resp = append(resp, mapSectionToResponse(sec))
	}
	return resp, nil
}

func (s *sectionService) GetTemplateSections(ctx context.Context, templateID string) ([]SectionResponse, error) {
	templateUUID, err := uuid.Parse(templateID)
	if err != nil {
		return nil, err
	}
	if _, err := s.repo.GetFormTemplateByID(ctx, templateUUID); err != nil {
		return nil, ErrTemplateNotFound
	}
	sections, err := s.repo.GetTemplateSectionsByTemplateID(ctx, pgtype.UUID{Bytes: templateUUID, Valid: true})
	if err != nil {
		return nil, err
	}
	resp := make([]SectionResponse, 0, len(sections))
	for _, sec := range sections {
		resp = append(resp, mapSectionToResponse(sec))
	}
	return resp, nil
}

func (s *sectionService) UpdateSection(ctx context.Context, sectionID string, dto UpdateSectionDto, userID string) (SectionResponse, error) {
	sectionUUID, err := uuid.Parse(sectionID)
	if err != nil {
		return SectionResponse{}, err
	}
	section, err := s.repo.GetFormSectionByID(ctx, sectionUUID)
	if err != nil {
		return SectionResponse{}, ErrSectionNotFound
	}
	// If it's a form section, verify ownership
	if section.FormID.Valid {
		if userID != "" {
			userUUID, err := uuid.Parse(userID)
			if err != nil {
				return SectionResponse{}, err
			}
			org, err := s.repo.GetOrganizationByUserID(ctx, userUUID)
			if err != nil {
				return SectionResponse{}, ErrAccessDenied
			}
			form, err := s.repo.GetFormByID(ctx, uuid.UUID(section.FormID.Bytes))
			if err != nil || form.OrganizationID != org.ID {
				return SectionResponse{}, ErrAccessDenied
			}
		}
	}
	// Template sections are global - no org check for now (any authenticated user can edit)
	updated, err := s.repo.UpdateFormSection(ctx, repo.UpdateFormSectionParams{
		ID:    sectionUUID,
		Title: dto.Title,
		Description: pgtype.Text{
			String: dto.Description,
			Valid:  dto.Description != "",
		},
		SortOrder: int32(dto.SortOrder),
	})
	if err != nil {
		return SectionResponse{}, err
	}
	return mapSectionToResponse(updated), nil
}

func (s *sectionService) DeleteSection(ctx context.Context, sectionID string, userID string) error {
	sectionUUID, err := uuid.Parse(sectionID)
	if err != nil {
		return err
	}
	section, err := s.repo.GetFormSectionByID(ctx, sectionUUID)
	if err != nil {
		return ErrSectionNotFound
	}
	if section.FormID.Valid {
		if userID != "" {
			userUUID, err := uuid.Parse(userID)
			if err != nil {
				return err
			}
			org, err := s.repo.GetOrganizationByUserID(ctx, userUUID)
			if err != nil {
				return ErrAccessDenied
			}
			form, err := s.repo.GetFormByID(ctx, uuid.UUID(section.FormID.Bytes))
			if err != nil || form.OrganizationID != org.ID {
				return ErrAccessDenied
			}
		}
	}
	return s.repo.DeleteFormSection(ctx, sectionUUID)
}

func mapSectionToResponse(s repo.FormSection) SectionResponse {
	formID := ""
	if s.FormID.Valid {
		formID = uuid.UUID(s.FormID.Bytes).String()
	}
	templateID := ""
	if s.TemplateID.Valid {
		templateID = uuid.UUID(s.TemplateID.Bytes).String()
	}
	return SectionResponse{
		ID:          s.ID.String(),
		FormID:      formID,
		TemplateID:  templateID,
		Title:       s.Title,
		Description: s.Description.String,
		SortOrder:   int(s.SortOrder),
		CreatedAt:   s.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   s.UpdatedAt.Time.Format(time.RFC3339),
	}
}
