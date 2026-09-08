package form_templates

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"preuvio/internal/repo"
	"preuvio/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var ErrTemplateFieldNotFound = errors.New("template field not found")
var ErrDuplicateKey = errors.New("field key already exists in this template")

type TemplateFieldService interface {
	CreateField(ctx context.Context, templateID string, dto CreateTemplateFieldDto) (TemplateFieldResponse, error)
	GetFieldsByTemplateID(ctx context.Context, templateID string) ([]TemplateFieldResponse, error)
	UpdateField(ctx context.Context, fieldID string, dto UpdateTemplateFieldDto) (TemplateFieldResponse, error)
	DeleteField(ctx context.Context, fieldID string) error
}

type templateFieldService struct {
	repo *repo.Queries
}

func NewTemplateFieldService(queries *repo.Queries) TemplateFieldService {
	return &templateFieldService{repo: queries}
}

// Delegates to unified logic - kept for backward compatibility, DTOs remain distinct for Swagger clarity
// but implementation is now single-source via shared mapper to avoid duplication.

func (s *templateFieldService) CreateField(ctx context.Context, templateID string, dto CreateTemplateFieldDto) (TemplateFieldResponse, error) {
	templateUUID, err := uuid.Parse(templateID)
	if err != nil {
		return TemplateFieldResponse{}, err
	}
	sectionUUID, err := uuid.Parse(dto.SectionID)
	if err != nil {
		return TemplateFieldResponse{}, err
	}
	if _, err := s.repo.GetFormTemplateByID(ctx, templateUUID); err != nil {
		return TemplateFieldResponse{}, ErrTemplateNotFound
	}
	var validationJSON []byte
	if dto.Validation != nil {
		validationJSON, err = json.Marshal(dto.Validation)
		if err != nil {
			return TemplateFieldResponse{}, err
		}
	}
	var optionsJSON []byte
	if dto.Options != nil {
		optionsJSON, err = json.Marshal(dto.Options)
		if err != nil {
			return TemplateFieldResponse{}, err
		}
	}
	section, err := s.repo.GetFormSectionByID(ctx, sectionUUID)
	if err != nil {
		return TemplateFieldResponse{}, ErrTemplateFieldNotFound
	}
	if !section.TemplateID.Valid || uuid.UUID(section.TemplateID.Bytes) != templateUUID {
		return TemplateFieldResponse{}, ErrTemplateFieldNotFound
	}
	// A lock: key is background, server-generated fld_8char. dto.Key ignored.
	var field repo.FormField
	created := false
	for i := 0; i < 3; i++ {
		key, err := utils.GenerateFieldKey()
		if err != nil {
			return TemplateFieldResponse{}, err
		}
		field, err = s.repo.CreateFormField(ctx, repo.CreateFormFieldParams{
			FormID: pgtype.UUID{Valid: false}, TemplateID: pgtype.UUID{Bytes: templateUUID, Valid: true},
			SectionID: sectionUUID, FieldType: dto.FieldType, Label: dto.Label, Key: key,
			Description: pgtype.Text{String: dto.Description, Valid: dto.Description != ""},
			Placeholder: pgtype.Text{String: dto.Placeholder, Valid: dto.Placeholder != ""},
			IsRequired: dto.IsRequired, SortOrder: int32(dto.SortOrder),
			Validation: validationJSON, Options: optionsJSON,
		})
		if err == nil {
			created = true
			break
		}
		if !utils.IsUniqueViolation(err) {
			return TemplateFieldResponse{}, err
		}
	}
	if !created {
		return TemplateFieldResponse{}, ErrDuplicateKey
	}
	return mapTemplateFieldToResponse(field), nil
}

func (s *templateFieldService) GetFieldsByTemplateID(ctx context.Context, templateID string) ([]TemplateFieldResponse, error) {
	templateUUID, err := uuid.Parse(templateID)
	if err != nil {
		return nil, err
	}
	fields, err := s.repo.GetTemplateFieldsByTemplateID(ctx, pgtype.UUID{Bytes: templateUUID, Valid: true})
	if err != nil {
		return nil, err
	}
	response := make([]TemplateFieldResponse, 0, len(fields))
	for _, f := range fields {
		response = append(response, mapTemplateFieldToResponse(f))
	}
	return response, nil
}

func (s *templateFieldService) UpdateField(ctx context.Context, fieldID string, dto UpdateTemplateFieldDto) (TemplateFieldResponse, error) {
	fieldUUID, err := uuid.Parse(fieldID)
	if err != nil {
		return TemplateFieldResponse{}, err
	}
	field, err := s.repo.GetFormFieldByID(ctx, fieldUUID)
	if err != nil {
		return TemplateFieldResponse{}, ErrTemplateFieldNotFound
	}
	if !field.TemplateID.Valid {
		return TemplateFieldResponse{}, ErrTemplateFieldNotFound
	}
	var validationJSON []byte
	if dto.Validation != nil {
		validationJSON, err = json.Marshal(dto.Validation)
		if err != nil {
			return TemplateFieldResponse{}, err
		}
	}
	var optionsJSON []byte
	if dto.Options != nil {
		optionsJSON, err = json.Marshal(dto.Options)
		if err != nil {
			return TemplateFieldResponse{}, err
		}
	}
	// A lock: key immutable after create. dto.Key ignored.
	updated, err := s.repo.UpdateFormField(ctx, repo.UpdateFormFieldParams{
		ID: fieldUUID, FieldType: dto.FieldType, Label: dto.Label,
		Description: pgtype.Text{String: dto.Description, Valid: dto.Description != ""},
		Placeholder: pgtype.Text{String: dto.Placeholder, Valid: dto.Placeholder != ""},
		IsRequired: dto.IsRequired, SortOrder: int32(dto.SortOrder),
		Validation: validationJSON, Options: optionsJSON,
	})
	if err != nil {
		return TemplateFieldResponse{}, err
	}
	return mapTemplateFieldToResponse(updated), nil
}

func (s *templateFieldService) DeleteField(ctx context.Context, fieldID string) error {
	fieldUUID, err := uuid.Parse(fieldID)
	if err != nil {
		return err
	}
	field, err := s.repo.GetFormFieldByID(ctx, fieldUUID)
	if err != nil {
		return ErrTemplateFieldNotFound
	}
	if !field.TemplateID.Valid {
		return ErrTemplateFieldNotFound
	}
	return s.repo.DeleteFormField(ctx, fieldUUID)
}

func mapTemplateFieldToResponse(f repo.FormField) TemplateFieldResponse {
	var validation interface{}
	if f.Validation != nil {
		_ = json.Unmarshal(f.Validation, &validation)
	}
	var options interface{}
	if f.Options != nil {
		_ = json.Unmarshal(f.Options, &options)
	}
	return TemplateFieldResponse{
		ID: f.ID.String(), TemplateID: uuid.UUID(f.TemplateID.Bytes).String(),
		SectionID: f.SectionID.String(), FieldType: f.FieldType, Label: f.Label, Key: f.Key,
		Description: f.Description.String, Placeholder: f.Placeholder.String,
		IsRequired: f.IsRequired, SortOrder: int(f.SortOrder),
		Validation: validation, Options: options,
		CreatedAt: f.CreatedAt.Time.Format(time.RFC3339), UpdatedAt: f.UpdatedAt.Time.Format(time.RFC3339),
	}
}
