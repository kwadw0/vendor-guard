package form_fields

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"preuvio/internal/repo"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var ErrFieldNotFound = errors.New("field not found")
var ErrAccessDenied = errors.New("access denied")
var ErrDuplicateKey = errors.New("field key already exists in this form")
var ErrTemplateNotFound = errors.New("template not found")
var ErrTemplateFieldNotFound = errors.New("template field not found")

type FormFieldService interface {
	CreateField(ctx context.Context, formID string, dto CreateFormFieldDto, userID string) (FormFieldResponse, error)
	GetFieldsByFormID(ctx context.Context, formID string, userID string) ([]FormFieldResponse, error)
	UpdateField(ctx context.Context, fieldID string, dto UpdateFormFieldDto, userID string) (FormFieldResponse, error)
	DeleteField(ctx context.Context, fieldID string, userID string) error
	// Template-scoped operations (shared implementation, same form_fields table with template_id)
	CreateTemplateField(ctx context.Context, templateID string, dto CreateFormFieldDto) (FormFieldResponse, error)
	GetTemplateFields(ctx context.Context, templateID string) ([]FormFieldResponse, error)
	UpdateTemplateField(ctx context.Context, fieldID string, dto UpdateFormFieldDto) (FormFieldResponse, error)
	DeleteTemplateField(ctx context.Context, fieldID string) error
}

type formFieldService struct {
	repo *repo.Queries
}

func NewService(queries *repo.Queries) FormFieldService {
	return &formFieldService{repo: queries}
}

func (s *formFieldService) CreateField(ctx context.Context, formID string, dto CreateFormFieldDto, userID string) (FormFieldResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return FormFieldResponse{}, err
	}

	formUUID, err := uuid.Parse(formID)
	if err != nil {
		return FormFieldResponse{}, err
	}

	sectionUUID, err := uuid.Parse(dto.SectionID)
	if err != nil {
		return FormFieldResponse{}, err
	}

	org, err := s.repo.GetOrganizationByUserID(ctx, userUUID)
	if err != nil {
		return FormFieldResponse{}, ErrAccessDenied
	}

	form, err := s.repo.GetFormByID(ctx, formUUID)
	if err != nil {
		return FormFieldResponse{}, ErrFieldNotFound
	}

	if form.OrganizationID != org.ID {
		return FormFieldResponse{}, ErrAccessDenied
	}

	var validationJSON []byte
	if dto.Validation != nil {
		validationJSON, err = json.Marshal(dto.Validation)
		if err != nil {
			return FormFieldResponse{}, err
		}
	}

	var optionsJSON []byte
	if dto.Options != nil {
		optionsJSON, err = json.Marshal(dto.Options)
		if err != nil {
			return FormFieldResponse{}, err
		}
	}

	section, err := s.repo.GetFormSectionByID(ctx, sectionUUID)
	if err != nil {
		return FormFieldResponse{}, ErrFieldNotFound
	}
	if !section.FormID.Valid || uuid.UUID(section.FormID.Bytes) != formUUID {
		return FormFieldResponse{}, ErrFieldNotFound
	}

	field, err := s.repo.CreateFormField(ctx, repo.CreateFormFieldParams{
		FormID:     pgtype.UUID{Bytes: formUUID, Valid: true},
		TemplateID: pgtype.UUID{Valid: false},
		SectionID:  sectionUUID,
		FieldType:  dto.FieldType,
		Label:      dto.Label,
		Key:        dto.Key,
		Description: pgtype.Text{
			String: dto.Description,
			Valid:  dto.Description != "",
		},
		Placeholder: pgtype.Text{
			String: dto.Placeholder,
			Valid:  dto.Placeholder != "",
		},
		IsRequired: dto.IsRequired,
		SortOrder:  int32(dto.SortOrder),
		Validation: validationJSON,
		Options:    optionsJSON,
	})
	if err != nil {
		return FormFieldResponse{}, err
	}
	return mapFieldToResponse(field), nil
}

func (s *formFieldService) GetFieldsByFormID(ctx context.Context, formID string, userID string) ([]FormFieldResponse, error) {
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
		return nil, ErrFieldNotFound
	}

	if form.OrganizationID != org.ID {
		return nil, ErrAccessDenied
	}

	fields, err := s.repo.GetFormFieldsByFormID(ctx, pgtype.UUID{Bytes: formUUID, Valid: true})
	if err != nil {
		return nil, err
	}

	response := make([]FormFieldResponse, 0, len(fields))
	for _, f := range fields {
		response = append(response, mapFieldToResponse(f))
	}
	return response, nil
}

func (s *formFieldService) UpdateField(ctx context.Context, fieldID string, dto UpdateFormFieldDto, userID string) (FormFieldResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return FormFieldResponse{}, err
	}

	fieldUUID, err := uuid.Parse(fieldID)
	if err != nil {
		return FormFieldResponse{}, err
	}

	org, err := s.repo.GetOrganizationByUserID(ctx, userUUID)
	if err != nil {
		return FormFieldResponse{}, ErrAccessDenied
	}

	field, err := s.repo.GetFormFieldByID(ctx, fieldUUID)
	if err != nil {
		return FormFieldResponse{}, ErrFieldNotFound
	}

	if !field.FormID.Valid {
		return FormFieldResponse{}, ErrFieldNotFound
	}

	form, err := s.repo.GetFormByID(ctx, uuid.UUID(field.FormID.Bytes))
	if err != nil {
		return FormFieldResponse{}, ErrFieldNotFound
	}

	if form.OrganizationID != org.ID {
		return FormFieldResponse{}, ErrAccessDenied
	}

	var validationJSON []byte
	if dto.Validation != nil {
		validationJSON, err = json.Marshal(dto.Validation)
		if err != nil {
			return FormFieldResponse{}, err
		}
	}

	var optionsJSON []byte
	if dto.Options != nil {
		optionsJSON, err = json.Marshal(dto.Options)
		if err != nil {
			return FormFieldResponse{}, err
		}
	}

	updated, err := s.repo.UpdateFormField(ctx, repo.UpdateFormFieldParams{
		ID: fieldUUID,
		FieldType: dto.FieldType,
		Label:     dto.Label,
		Key:       dto.Key,
		Description: pgtype.Text{
			String: dto.Description,
			Valid:  dto.Description != "",
		},
		Placeholder: pgtype.Text{
			String: dto.Placeholder,
			Valid:  dto.Placeholder != "",
		},
		IsRequired: dto.IsRequired,
		SortOrder:  int32(dto.SortOrder),
		Validation: validationJSON,
		Options:    optionsJSON,
	})
	if err != nil {
		return FormFieldResponse{}, err
	}
	return mapFieldToResponse(updated), nil
}

func (s *formFieldService) DeleteField(ctx context.Context, fieldID string, userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	fieldUUID, err := uuid.Parse(fieldID)
	if err != nil {
		return err
	}

	org, err := s.repo.GetOrganizationByUserID(ctx, userUUID)
	if err != nil {
		return ErrAccessDenied
	}

	field, err := s.repo.GetFormFieldByID(ctx, fieldUUID)
	if err != nil {
		return ErrFieldNotFound
	}

	if !field.FormID.Valid {
		return ErrFieldNotFound
	}

	form, err := s.repo.GetFormByID(ctx, uuid.UUID(field.FormID.Bytes))
	if err != nil {
		return ErrFieldNotFound
	}

	if form.OrganizationID != org.ID {
		return ErrAccessDenied
	}

	return s.repo.DeleteFormField(ctx, fieldUUID)
}

func mapFieldToResponse(f repo.FormField) FormFieldResponse {
	var validation interface{}
	if f.Validation != nil {
		_ = json.Unmarshal(f.Validation, &validation)
	}

	var options interface{}
	if f.Options != nil {
		_ = json.Unmarshal(f.Options, &options)
	}

	formID := ""
	if f.FormID.Valid {
		formID = uuid.UUID(f.FormID.Bytes).String()
	}
	templateID := ""
	if f.TemplateID.Valid {
		templateID = uuid.UUID(f.TemplateID.Bytes).String()
	}
	// Prefer formID, fallback to templateID for unified response
	ownerID := formID
	if ownerID == "" {
		ownerID = templateID
	}

	return FormFieldResponse{
		ID:          f.ID.String(),
		FormID:      ownerID,
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

func (s *formFieldService) CreateTemplateField(ctx context.Context, templateID string, dto CreateFormFieldDto) (FormFieldResponse, error) {
	templateUUID, err := uuid.Parse(templateID)
	if err != nil {
		return FormFieldResponse{}, err
	}
	sectionUUID, err := uuid.Parse(dto.SectionID)
	if err != nil {
		return FormFieldResponse{}, err
	}
	if _, err := s.repo.GetFormTemplateByID(ctx, templateUUID); err != nil {
		return FormFieldResponse{}, ErrTemplateNotFound
	}
	var validationJSON []byte
	if dto.Validation != nil {
		validationJSON, err = json.Marshal(dto.Validation)
		if err != nil {
			return FormFieldResponse{}, err
		}
	}
	var optionsJSON []byte
	if dto.Options != nil {
		optionsJSON, err = json.Marshal(dto.Options)
		if err != nil {
			return FormFieldResponse{}, err
		}
	}
	section, err := s.repo.GetFormSectionByID(ctx, sectionUUID)
	if err != nil {
		return FormFieldResponse{}, ErrTemplateFieldNotFound
	}
	if !section.TemplateID.Valid || uuid.UUID(section.TemplateID.Bytes) != templateUUID {
		return FormFieldResponse{}, ErrTemplateFieldNotFound
	}
	field, err := s.repo.CreateFormField(ctx, repo.CreateFormFieldParams{
		FormID:     pgtype.UUID{Valid: false},
		TemplateID: pgtype.UUID{Bytes: templateUUID, Valid: true},
		SectionID:  sectionUUID,
		FieldType:  dto.FieldType,
		Label:      dto.Label,
		Key:        dto.Key,
		Description: pgtype.Text{String: dto.Description, Valid: dto.Description != ""},
		Placeholder: pgtype.Text{String: dto.Placeholder, Valid: dto.Placeholder != ""},
		IsRequired: dto.IsRequired,
		SortOrder:  int32(dto.SortOrder),
		Validation: validationJSON,
		Options:    optionsJSON,
	})
	if err != nil {
		return FormFieldResponse{}, err
	}
	return mapFieldToResponse(field), nil
}

func (s *formFieldService) GetTemplateFields(ctx context.Context, templateID string) ([]FormFieldResponse, error) {
	templateUUID, err := uuid.Parse(templateID)
	if err != nil {
		return nil, err
	}
	fields, err := s.repo.GetTemplateFieldsByTemplateID(ctx, pgtype.UUID{Bytes: templateUUID, Valid: true})
	if err != nil {
		return nil, err
	}
	response := make([]FormFieldResponse, 0, len(fields))
	for _, f := range fields {
		response = append(response, mapFieldToResponse(f))
	}
	return response, nil
}

func (s *formFieldService) UpdateTemplateField(ctx context.Context, fieldID string, dto UpdateFormFieldDto) (FormFieldResponse, error) {
	fieldUUID, err := uuid.Parse(fieldID)
	if err != nil {
		return FormFieldResponse{}, err
	}
	field, err := s.repo.GetFormFieldByID(ctx, fieldUUID)
	if err != nil {
		return FormFieldResponse{}, ErrTemplateFieldNotFound
	}
	if !field.TemplateID.Valid {
		return FormFieldResponse{}, ErrTemplateFieldNotFound
	}
	var validationJSON []byte
	if dto.Validation != nil {
		validationJSON, err = json.Marshal(dto.Validation)
		if err != nil {
			return FormFieldResponse{}, err
		}
	}
	var optionsJSON []byte
	if dto.Options != nil {
		optionsJSON, err = json.Marshal(dto.Options)
		if err != nil {
			return FormFieldResponse{}, err
		}
	}
	updated, err := s.repo.UpdateFormField(ctx, repo.UpdateFormFieldParams{
		ID: fieldUUID, FieldType: dto.FieldType, Label: dto.Label, Key: dto.Key,
		Description: pgtype.Text{String: dto.Description, Valid: dto.Description != ""},
		Placeholder: pgtype.Text{String: dto.Placeholder, Valid: dto.Placeholder != ""},
		IsRequired: dto.IsRequired, SortOrder: int32(dto.SortOrder),
		Validation: validationJSON, Options: optionsJSON,
	})
	if err != nil {
		return FormFieldResponse{}, err
	}
	return mapFieldToResponse(updated), nil
}

func (s *formFieldService) DeleteTemplateField(ctx context.Context, fieldID string) error {
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
