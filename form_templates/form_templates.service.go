package form_templates

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"preuvio/internal/repo"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrTemplateNotFound = errors.New("template not found")
var ErrAccessDenied = errors.New("access denied")

type FormTemplateService interface {
	CreateTemplate(ctx context.Context, dto CreateFormTemplateDto) (FormTemplateResponse, error)
	GetTemplateByID(ctx context.Context, id string) (FormTemplateResponse, error)
	GetAllTemplates(ctx context.Context) ([]FormTemplateResponse, error)
	UpdateTemplate(ctx context.Context, id string, dto UpdateFormTemplateDto) (FormTemplateResponse, error)
	DeleteTemplate(ctx context.Context, id string) error
	CloneTemplateToForm(ctx context.Context, templateID string, dto CloneTemplateDto, userID string) (CloneFormResponse, error)
	GetTemplateDetail(ctx context.Context, id string) (TemplateDetailResponse, error)
}

type formTemplateService struct {
	repo *repo.Queries
	pool *pgxpool.Pool
}

func NewService(queries *repo.Queries) FormTemplateService {
	return &formTemplateService{repo: queries}
}

func NewServiceWithPool(pool *pgxpool.Pool, queries *repo.Queries) FormTemplateService {
	return &formTemplateService{repo: queries, pool: pool}
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

func (s *formTemplateService) CloneTemplateToForm(ctx context.Context, templateID string, dto CloneTemplateDto, userID string) (CloneFormResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return CloneFormResponse{}, err
	}

	org, err := s.repo.GetOrganizationByUserID(ctx, userUUID)
	if err != nil {
		return CloneFormResponse{}, ErrAccessDenied
	}

	templateUUID, err := uuid.Parse(templateID)
	if err != nil {
		return CloneFormResponse{}, err
	}

	// Verify template exists
	if _, err := s.repo.GetFormTemplateByID(ctx, templateUUID); err != nil {
		return CloneFormResponse{}, ErrTemplateNotFound
	}

	// Transactional deep-copy: form + sections + fields atomically
	if s.pool == nil {
		return CloneFormResponse{}, errors.New("clone requires database pool - use NewServiceWithPool")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return CloneFormResponse{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := s.repo.WithTx(tx)

	form, err := qtx.CloneTemplateToForm(ctx, repo.CloneTemplateToFormParams{
		OrganizationID: org.ID,
		TemplateID:     pgtype.UUID{Bytes: templateUUID, Valid: true},
		Title:          dto.Title,
	})
	if err != nil {
		return CloneFormResponse{}, err
	}

	// Clone sections with fresh IDs and remap
	templateSections, err := qtx.GetTemplateSectionsByTemplateID(ctx, pgtype.UUID{Bytes: templateUUID, Valid: true})
	if err != nil {
		return CloneFormResponse{}, err
	}

	sectionMap := make(map[uuid.UUID]uuid.UUID, len(templateSections))
	for _, ts := range templateSections {
		ns, err := qtx.CreateFormSection(ctx, repo.CreateFormSectionParams{
			FormID:      pgtype.UUID{Bytes: form.ID, Valid: true},
			TemplateID:  pgtype.UUID{Valid: false},
			Title:       ts.Title,
			Description: ts.Description,
			SortOrder:   ts.SortOrder,
		})
		if err != nil {
			return CloneFormResponse{}, err
		}
		sectionMap[ts.ID] = ns.ID
	}

	// Clone fields with remapped section IDs
	templateFields, err := qtx.GetTemplateFieldsByTemplateID(ctx, pgtype.UUID{Bytes: templateUUID, Valid: true})
	if err != nil {
		return CloneFormResponse{}, err
	}

	for _, tf := range templateFields {
		newSectionID, ok := sectionMap[tf.SectionID]
		if !ok {
			// Section missing — skip orphan field (should not happen if data is consistent)
			continue
		}
		_, err := qtx.CreateFormField(ctx, repo.CreateFormFieldParams{
			FormID:      pgtype.UUID{Bytes: form.ID, Valid: true},
			TemplateID:  pgtype.UUID{Valid: false},
			SectionID:   newSectionID,
			FieldType:   tf.FieldType,
			Label:       tf.Label,
			Key:         tf.Key,
			Description: tf.Description,
			Placeholder: tf.Placeholder,
			IsRequired:  tf.IsRequired,
			SortOrder:   tf.SortOrder,
			Validation:  tf.Validation,
			Options:     tf.Options,
		})
		if err != nil {
			return CloneFormResponse{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return CloneFormResponse{}, err
	}

	return mapFormToCloneResponse(form), nil
}

func (s *formTemplateService) GetTemplateDetail(ctx context.Context, id string) (TemplateDetailResponse, error) {
	templateUUID, err := uuid.Parse(id)
	if err != nil {
		return TemplateDetailResponse{}, err
	}
	template, err := s.repo.GetFormTemplateByID(ctx, templateUUID)
	if err != nil {
		return TemplateDetailResponse{}, ErrTemplateNotFound
	}
	sections, err := s.repo.GetTemplateSectionsByTemplateID(ctx, pgtype.UUID{Bytes: templateUUID, Valid: true})
	if err != nil {
		return TemplateDetailResponse{}, err
	}
	fields, err := s.repo.GetTemplateFieldsByTemplateID(ctx, pgtype.UUID{Bytes: templateUUID, Valid: true})
	if err != nil {
		return TemplateDetailResponse{}, err
	}
	// Group fields by section
	fieldsBySection := make(map[uuid.UUID][]TemplateFieldDetail, len(sections))
	for _, f := range fields {
		fieldsBySection[f.SectionID] = append(fieldsBySection[f.SectionID], mapTemplateFieldToDetail(f))
	}
	detailSections := make([]TemplateSectionWithFields, 0, len(sections))
	for _, sec := range sections {
		detailSections = append(detailSections, TemplateSectionWithFields{
			Section: mapSectionToDetail(sec),
			Fields:  fieldsBySection[sec.ID],
		})
	}
	// Ensure empty slice not null
	for i := range detailSections {
		if detailSections[i].Fields == nil {
			detailSections[i].Fields = []TemplateFieldDetail{}
		}
	}
	return TemplateDetailResponse{
		Template: mapTemplateToResponse(template),
		Sections: detailSections,
	}, nil
}

func mapSectionToDetail(s repo.FormSection) SectionDetailResponse {
	templateID := ""
	if s.TemplateID.Valid {
		templateID = uuid.UUID(s.TemplateID.Bytes).String()
	}
	return SectionDetailResponse{
		ID:          s.ID.String(),
		TemplateID:  templateID,
		Title:       s.Title,
		Description: s.Description.String,
		SortOrder:   int(s.SortOrder),
		CreatedAt:   s.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   s.UpdatedAt.Time.Format(time.RFC3339),
	}
}

func mapTemplateFieldToDetail(f repo.FormField) TemplateFieldDetail {
	var validation interface{}
	if f.Validation != nil {
		_ = json.Unmarshal(f.Validation, &validation)
	}
	var options interface{}
	if f.Options != nil {
		_ = json.Unmarshal(f.Options, &options)
	}
	return TemplateFieldDetail{
		ID:          f.ID.String(),
		TemplateID:  uuid.UUID(f.TemplateID.Bytes).String(),
		SectionID:   f.SectionID.String(),
		FieldType:   f.FieldType,
		Label:       f.Label,
		Key:         f.Key,
		Description: f.Description.String,
		Placeholder: f.Placeholder.String,
		IsRequired:  f.IsRequired,
		SortOrder:   int(f.SortOrder),
		Validation:  validation,
		Options:     options,
		CreatedAt:   f.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   f.UpdatedAt.Time.Format(time.RFC3339),
	}
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

func mapFormToCloneResponse(f repo.Form) CloneFormResponse {
	templateID := ""
	if f.TemplateID.Valid {
		templateID = uuid.UUID(f.TemplateID.Bytes).String()
	}
	return CloneFormResponse{
		ID:             f.ID.String(),
		OrganizationID: f.OrganizationID.String(),
		TemplateID:     templateID,
		Title:          f.Title,
		Description:    f.Description.String,
		Status:         f.Status,
		CreatedAt:      f.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:      f.UpdatedAt.Time.Format(time.RFC3339),
	}
}
