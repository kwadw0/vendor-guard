package organizations

import (
	"context"
	"errors"

	"preuvio/internal/repo"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var ErrOrganizationNotFound = errors.New("organization not found")
var ErrAccessDenied = errors.New("access denied")

type OrganizationService interface {
	CreateOrganization(ctx context.Context, dto CreateOrganizationDto, userID string) (OrganizationResponseDto, error)
	GetOrganizationById(ctx context.Context, id uuid.UUID) (OrganizationResponseDto, error)
	GetOrganizationByUserID(ctx context.Context, userID string) (OrganizationResponseDto, error)
	GetAllOrganizations(ctx context.Context) ([]OrganizationResponseDto, error)
	UpdateOrganization(ctx context.Context, id uuid.UUID, dto UpdateOrganizationDto) (OrganizationResponseDto, error)
	DeleteOrganization(ctx context.Context, id uuid.UUID) error
	VerifyOrgMember(ctx context.Context, userID string, orgID uuid.UUID) error
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
	existing, err := s.queries.GetOrganizationById(ctx, id)
	if err != nil {
		return OrganizationResponseDto{}, ErrOrganizationNotFound
	}

	// Partial update: nil fields keep existing values (PATCH semantics).
	name := existing.Name
	if dto.Name != nil {
		name = *dto.Name
	}
	description := existing.Description.String
	if dto.Description != nil {
		description = *dto.Description
	}
	websiteURL := existing.WebsiteUrl.String
	if dto.WebsiteURL != nil {
		websiteURL = *dto.WebsiteURL
	}
	industry := existing.Industry.String
	if dto.Industry != nil {
		industry = *dto.Industry
	}
	teamSize := existing.TeamSize.String
	if dto.TeamSize != nil {
		teamSize = *dto.TeamSize
	}
	primaryCustomerType := existing.PrimaryCustomerType.String
	if dto.PrimaryCustomerType != nil {
		primaryCustomerType = *dto.PrimaryCustomerType
	}
	ownerRole := existing.OwnerRole
	if dto.OwnerRole != nil {
		ownerRole = *dto.OwnerRole
	}

	org, err := s.queries.UpdateOrganization(ctx, repo.UpdateOrganizationParams{
		ID:                  id,
		Name:                name,
		Description:         toPgText(description),
		WebsiteUrl:          toPgText(websiteURL),
		Industry:            toPgText(industry),
		TeamSize:            toPgText(teamSize),
		PrimaryCustomerType: toPgText(primaryCustomerType),
		OwnerRole:           ownerRole,
	})
	if err != nil {
		return OrganizationResponseDto{}, err
	}
	return mapToDto(org), nil
}

func (s *organizationService) DeleteOrganization(ctx context.Context, id uuid.UUID) error {
	if _, err := s.queries.GetOrganizationById(ctx, id); err != nil {
		return ErrOrganizationNotFound
	}
	return s.queries.DeleteOrganization(ctx, id)
}

// VerifyOrgMember returns the org if the user belongs to it, else ErrAccessDenied.
// Used by handlers to gate writes to the caller's own organization.
func (s *organizationService) VerifyOrgMember(ctx context.Context, userID string, orgID uuid.UUID) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return ErrAccessDenied
	}
	org, err := s.queries.GetOrganizationByUserID(ctx, uid)
	if err != nil {
		return ErrAccessDenied
	}
	if org.ID != orgID {
		return ErrAccessDenied
	}
	return nil
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
