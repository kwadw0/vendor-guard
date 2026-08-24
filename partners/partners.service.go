package partners

import (
	"context"
	"errors"
	"time"

	"preuvio/internal/repo"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var ErrAccessDenied = errors.New("access denied")

type PartnerService interface {
	CreatePartner(ctx context.Context, dto CreatePartnerDto, userID string) (PartnerResponse, error)
	GetPartnerByID(ctx context.Context, id, userID string) (PartnerResponse, error)
	GetAllPartners(ctx context.Context, userID string) ([]PartnerResponse, error)
	UpdatePartner(ctx context.Context, dto UpdatePartnerDto, userID string) (PartnerResponse, error)
	DeletePartner(ctx context.Context, id, userID string) error
}

type partnerService struct {
	repo *repo.Queries
}

func NewService(partnerRepo *repo.Queries) PartnerService {
	return &partnerService{repo: partnerRepo}
}

func (s *partnerService) CreatePartner(
	ctx context.Context,
	dto CreatePartnerDto,
	userID string,
) (PartnerResponse, error) {

	id, err := uuid.Parse(userID)
	if err != nil {
		return PartnerResponse{}, err
	}

	// Verify user is an internal organization member
	org, err := s.repo.GetOrganizationByUserID(ctx, id)
	if err != nil {
		return PartnerResponse{}, ErrAccessDenied
	}

	partner, err := s.repo.CreatePartners(ctx, repo.CreatePartnersParams{
		OrganizationID: org.ID,
		Name:           dto.Name,
		Email:          dto.Email,
		Phone: pgtype.Text{
			String: dto.Phone,
			Valid:  dto.Phone != "",
		},
	})
	if err != nil {
		return PartnerResponse{}, err
	}

	return mapPartnerToResponse(partner), nil
}

func (s *partnerService) GetPartnerByID(
	ctx context.Context,
	id, userID string,
) (PartnerResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return PartnerResponse{}, err
	}

	partnerUUID, err := uuid.Parse(id)
	if err != nil {
		return PartnerResponse{}, err
	}

	// Check if user is an internal org member
	org, orgErr := s.repo.GetOrganizationByUserID(ctx, userUUID)
	if orgErr == nil {
		// Internal user: verify partner belongs to their org
		partner, err := s.repo.GetPartnerById(ctx, partnerUUID)
		if err != nil {
			return PartnerResponse{}, err
		}
		if partner.OrganizationID != org.ID {
			return PartnerResponse{}, ErrAccessDenied
		}
		return mapPartnerToResponse(partner), nil
	}

	// Check if user is a partner user
	partnerUser, pErr := s.repo.GetPartnerByUserID(ctx, userUUID)
	if pErr == nil {
		// Partner user: only can see their own partner
		if partnerUser.ID != partnerUUID {
			return PartnerResponse{}, ErrAccessDenied
		}
		return mapPartnerToResponse(partnerUser), nil
	}

	return PartnerResponse{}, ErrAccessDenied
}

func (s *partnerService) GetAllPartners(
	ctx context.Context,
	userID string,
) ([]PartnerResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	// Check if user is an internal org member
	org, orgErr := s.repo.GetOrganizationByUserID(ctx, userUUID)
	if orgErr == nil {
		partners, err := s.repo.GetPartnersByOrg(ctx, org.ID)
		if err != nil {
			return nil, err
		}
		response := make([]PartnerResponse, 0, len(partners))
		for _, p := range partners {
			response = append(response, mapPartnerToResponse(p))
		}
		return response, nil
	}

	// Check if user is a partner user
	partnerUser, pErr := s.repo.GetPartnerByUserID(ctx, userUUID)
	if pErr == nil {
		return []PartnerResponse{mapPartnerToResponse(partnerUser)}, nil
	}

	return nil, ErrAccessDenied
}

func (s *partnerService) UpdatePartner(
	ctx context.Context,
	dto UpdatePartnerDto,
	userID string,
) (PartnerResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return PartnerResponse{}, err
	}

	// Only internal org members can update partners
	org, err := s.repo.GetOrganizationByUserID(ctx, id)
	if err != nil {
		return PartnerResponse{}, ErrAccessDenied
	}

	// Verify partner belongs to user's org
	partnerUUID, err := uuid.Parse(dto.ID)
	if err != nil {
		return PartnerResponse{}, err
	}
	partner, err := s.repo.GetPartnerById(ctx, partnerUUID)
	if err != nil {
		return PartnerResponse{}, err
	}
	if partner.OrganizationID != org.ID {
		return PartnerResponse{}, ErrAccessDenied
	}

	partner, err = s.repo.UpdatePartner(ctx, repo.UpdatePartnerParams{
		ID: partnerUUID,
		Name:  dto.Name,
		Email: dto.Email,
		Phone: pgtype.Text{
			String: dto.Phone,
			Valid:  dto.Phone != "",
		},
	})
	if err != nil {
		return PartnerResponse{}, err
	}
	return mapPartnerToResponse(partner), nil
}

func (s *partnerService) DeletePartner(ctx context.Context, id, userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	// Only internal org members can delete partners
	org, err := s.repo.GetOrganizationByUserID(ctx, userUUID)
	if err != nil {
		return ErrAccessDenied
	}

	// Verify partner belongs to user's org
	partnerUUID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	partner, err := s.repo.GetPartnerById(ctx, partnerUUID)
	if err != nil {
		return err
	}
	if partner.OrganizationID != org.ID {
		return ErrAccessDenied
	}

	return s.repo.DeletePartner(ctx, partnerUUID)
}

func mapPartnerToResponse(p repo.Partner) PartnerResponse {
	return PartnerResponse{
		ID:             p.ID.String(),
		OrganizationID: p.OrganizationID.String(),
		Name:           p.Name,
		Email:          p.Email,
		Phone:          p.Phone.String,
		Status:         string(p.Status.PartnerStatus),
		CreatedAt:      p.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:      p.UpdatedAt.Time.Format(time.RFC3339),
	}
}
