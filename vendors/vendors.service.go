package vendors

import (
	"context"
	"errors"
	"time"

	"vendor-guard/internal/repo"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var ErrAccessDenied = errors.New("access denied")

type VendorService interface {
	CreateVendor(ctx context.Context, dto CreateVendorDto, userID string) (VendorResponse, error)
	GetVendorByID(ctx context.Context, id, userID string) (VendorResponse, error)
	GetAllVendors(ctx context.Context, userID string) ([]VendorResponse, error)
	UpdateVendor(ctx context.Context, dto UpdateVendorDto, userID string) (VendorResponse, error)
	DeleteVendor(ctx context.Context, id, userID string) error
}

type vendorService struct {
	repo *repo.Queries
}

func NewService(vendorRepo *repo.Queries) VendorService {
	return &vendorService{repo: vendorRepo}
}

func (s *vendorService) CreateVendor(
	ctx context.Context,
	dto CreateVendorDto,
	userID string,
) (VendorResponse, error) {

	id, err := uuid.Parse(userID)
	if err != nil {
		return VendorResponse{}, err
	}

	// Verify user is an internal organization member
	org, err := s.repo.GetOrganizationByUserID(ctx, id)
	if err != nil {
		return VendorResponse{}, ErrAccessDenied
	}

	vendor, err := s.repo.CreateVendors(ctx, repo.CreateVendorsParams{
		OrganizationID: org.ID,
		Name:           dto.Name,
		Email:          dto.Email,
		Phone: pgtype.Text{
			String: dto.Phone,
			Valid:  dto.Phone != "",
		},
	})
	if err != nil {
		return VendorResponse{}, err
	}

	return mapVendorToResponse(vendor), nil
}

func (s *vendorService) GetVendorByID(
	ctx context.Context,
	id, userID string,
) (VendorResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return VendorResponse{}, err
	}

	vendorUUID, err := uuid.Parse(id)
	if err != nil {
		return VendorResponse{}, err
	}

	// Check if user is an internal org member
	org, orgErr := s.repo.GetOrganizationByUserID(ctx, userUUID)
	if orgErr == nil {
		// Internal user: verify vendor belongs to their org
		vendor, err := s.repo.GetVendorById(ctx, vendorUUID)
		if err != nil {
			return VendorResponse{}, err
		}
		if vendor.OrganizationID != org.ID {
			return VendorResponse{}, ErrAccessDenied
		}
		return mapVendorToResponse(vendor), nil
	}

	// Check if user is a vendor user
	vendorUser, vErr := s.repo.GetVendorByUserID(ctx, userUUID)
	if vErr == nil {
		// Vendor user: only can see their own vendor
		if vendorUser.ID != vendorUUID {
			return VendorResponse{}, ErrAccessDenied
		}
		return mapVendorToResponse(vendorUser), nil
	}

	return VendorResponse{}, ErrAccessDenied
}

func (s *vendorService) GetAllVendors(
	ctx context.Context,
	userID string,
) ([]VendorResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	// Check if user is an internal org member
	org, orgErr := s.repo.GetOrganizationByUserID(ctx, userUUID)
	if orgErr == nil {
		vendors, err := s.repo.GetVendorsByOrg(ctx, org.ID)
		if err != nil {
			return nil, err
		}
		response := make([]VendorResponse, 0, len(vendors))
		for _, v := range vendors {
			response = append(response, mapVendorToResponse(v))
		}
		return response, nil
	}

	// Check if user is a vendor user
	vendorUser, vErr := s.repo.GetVendorByUserID(ctx, userUUID)
	if vErr == nil {
		return []VendorResponse{mapVendorToResponse(vendorUser)}, nil
	}

	return nil, ErrAccessDenied
}

func (s *vendorService) UpdateVendor(
	ctx context.Context,
	dto UpdateVendorDto,
	userID string,
) (VendorResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return VendorResponse{}, err
	}

	// Only internal org members can update vendors
	org, err := s.repo.GetOrganizationByUserID(ctx, id)
	if err != nil {
		return VendorResponse{}, ErrAccessDenied
	}

	// Verify vendor belongs to user's org
	vendorUUID, err := uuid.Parse(dto.ID)
	if err != nil {
		return VendorResponse{}, err
	}
	vendor, err := s.repo.GetVendorById(ctx, vendorUUID)
	if err != nil {
		return VendorResponse{}, err
	}
	if vendor.OrganizationID != org.ID {
		return VendorResponse{}, ErrAccessDenied
	}

	vendor, err = s.repo.UpdateVendor(ctx, repo.UpdateVendorParams{
		ID: vendorUUID,
		Name:  dto.Name,
		Email: dto.Email,
		Phone: pgtype.Text{
			String: dto.Phone,
			Valid:  dto.Phone != "",
		},
	})
	if err != nil {
		return VendorResponse{}, err
	}
	return mapVendorToResponse(vendor), nil
}

func (s *vendorService) DeleteVendor(ctx context.Context, id, userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	// Only internal org members can delete vendors
	org, err := s.repo.GetOrganizationByUserID(ctx, userUUID)
	if err != nil {
		return ErrAccessDenied
	}

	// Verify vendor belongs to user's org
	vendorUUID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	vendor, err := s.repo.GetVendorById(ctx, vendorUUID)
	if err != nil {
		return err
	}
	if vendor.OrganizationID != org.ID {
		return ErrAccessDenied
	}

	return s.repo.DeleteVendor(ctx, vendorUUID)
}

func mapVendorToResponse(v repo.Vendor) VendorResponse {
	return VendorResponse{
		ID:             v.ID.String(),
		OrganizationID: v.OrganizationID.String(),
		Name:           v.Name,
		Email:          v.Email,
		Phone:          v.Phone.String,
		Status:         string(v.Status.VendorStatus),
		CreatedAt:      v.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:      v.UpdatedAt.Time.Format(time.RFC3339),
	}
}
