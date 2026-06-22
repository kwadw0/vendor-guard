package organizations

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"vendor-guard/internal/repo"
)

var ErrOrganizationNotFound = errors.New("organization not found")

type OrganizationService interface {
	CreateOrganization(ctx context.Context, dto CreateOrganizationDto, userID string) (OrganizationResponseDto, error)
	GetOrganizationById(ctx context.Context, id uuid.UUID) (OrganizationResponseDto, error)
	GetOrganizationByUserID(ctx context.Context, userID string) (OrganizationResponseDto, error)
	GetAllOrganizations(ctx context.Context) ([]OrganizationResponseDto, error)
	UpdateOrganization(ctx context.Context, id uuid.UUID, dto UpdateOrganizationDto) (OrganizationResponseDto, error)
	DeleteOrganization(ctx context.Context, id uuid.UUID) error
}

type organizationService struct {
	queries *repo.Queries
}

func NewOrganizationService(queries *repo.Queries) OrganizationService {
	return &organizationService{queries: queries}
}

func (s *organizationService) CreateOrganization(ctx context.Context, dto CreateOrganizationDto, userID string) (OrganizationResponseDto, error) {
	// Parse the user ID coming from the JWT claim (always a string).
	uid, err := uuid.Parse(userID)
	if err != nil {
		return OrganizationResponseDto{}, errors.New("invalid user id")
	}

	// 1. Create the organization.
	org, err := s.queries.CreateOrganization(ctx, repo.CreateOrganizationParams{
		Name:                dto.Name,
		Description:         toPgText(dto.Description),
		WebsiteUrl:          toPgText(dto.WebsiteURL),
		Industry:            toPgText(dto.Industry),
		TeamSize:            toPgText(dto.TeamSize),
		PrimaryCustomerType: toPgText(dto.PrimaryCustomerType),
		OwnerRole:           dto.OwnerRole,
	})
	if err != nil {
		return OrganizationResponseDto{}, err
	}

	// 2. Link the authenticated user to the newly created organization.
	_, err = s.queries.UpdateUserOrganization(ctx, repo.UpdateUserOrganizationParams{
		ID:             uid,
		OrganizationID: pgtype.UUID{Bytes: org.ID, Valid: true},
	})
	if err != nil {
		return OrganizationResponseDto{}, err
	}

	return mapToDto(org), nil
}

func (s *organizationService) GetOrganizationById(ctx context.Context, id uuid.UUID) (OrganizationResponseDto, error) {
	org, err := s.queries.GetOrganizationById(ctx, id)
	if err != nil {
		return OrganizationResponseDto{}, ErrOrganizationNotFound
	}
	return mapToDto(org), nil
}

func (s *organizationService) GetOrganizationByUserID(ctx context.Context, userID string) (OrganizationResponseDto, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return OrganizationResponseDto{}, errors.New("invalid user id")
	}

	org, err := s.queries.GetOrganizationByUserID(ctx, uid)
	if err != nil {
		return OrganizationResponseDto{}, ErrOrganizationNotFound
	}
	return mapToDto(org), nil
}

func (s *organizationService) GetAllOrganizations(ctx context.Context) ([]OrganizationResponseDto, error) {
	orgs, err := s.queries.GetAllOrganizations(ctx)
	if err != nil {
		return nil, err
	}

	dtos := make([]OrganizationResponseDto, len(orgs))
	for i, org := range orgs {
		dtos[i] = mapToDto(org)
	}
	return dtos, nil
}

func (s *organizationService) UpdateOrganization(ctx context.Context, id uuid.UUID, dto UpdateOrganizationDto) (OrganizationResponseDto, error) {
	org, err := s.queries.UpdateOrganization(ctx, repo.UpdateOrganizationParams{
		ID:                  id,
		Name:                dto.Name,
		Description:         toPgText(dto.Description),
		WebsiteUrl:          toPgText(dto.WebsiteURL),
		Industry:            toPgText(dto.Industry),
		TeamSize:            toPgText(dto.TeamSize),
		PrimaryCustomerType: toPgText(dto.PrimaryCustomerType),
		OwnerRole:           dto.OwnerRole,
	})
	if err != nil {
		return OrganizationResponseDto{}, err
	}
	return mapToDto(org), nil
}

func (s *organizationService) DeleteOrganization(ctx context.Context, id uuid.UUID) error {
	return s.queries.DeleteOrganization(ctx, id)
}

// --- helpers ---

func toPgText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

func mapToDto(org repo.Organization) OrganizationResponseDto {
	return OrganizationResponseDto{
		ID:                  org.ID,
		Name:                org.Name,
		Description:         org.Description.String,
		WebsiteURL:          org.WebsiteUrl.String,
		Industry:            org.Industry.String,
		TeamSize:            org.TeamSize.String,
		PrimaryCustomerType: org.PrimaryCustomerType.String,
		OwnerRole:           org.OwnerRole,
		IsActive:            org.IsActive,
		CreatedAt:           org.CreatedAt.Time,
		UpdatedAt:           org.UpdatedAt.Time,
	}
}
