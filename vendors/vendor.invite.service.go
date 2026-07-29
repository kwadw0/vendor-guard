package vendors

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"vendor-guard/auth/jwt"
	"vendor-guard/internal/repo"
	"vendor-guard/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrInvitationNotFound    = errors.New("invitation not found or expired")
	ErrInvitationExpired     = errors.New("invitation has expired")
	ErrInvitationAlreadyUsed = errors.New("invitation already accepted")
	ErrVendorMismatch        = errors.New("vendor does not belong to your organization")
	ErrVendorUserExists      = errors.New("a user with this email already belongs to an organization")
)

type InviteService interface {
	InviteUser(ctx context.Context, vendorID, invitedByUserID string, dto InviteVendorUserDto) (InvitationResponse, error)
	AcceptInvitation(ctx context.Context, dto AcceptInviteDto) (*TokenResponse, error)
	GetInvitationsByVendor(ctx context.Context, vendorID, userID string) ([]InvitationResponse, error)
}

type inviteService struct {
	repo      *repo.Queries
	jwtSecret string
}

func NewInviteService(r *repo.Queries, jwtSecret string) InviteService {
	return &inviteService{repo: r, jwtSecret: jwtSecret}
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *inviteService) InviteUser(ctx context.Context, vendorID, invitedByUserID string, dto InviteVendorUserDto) (InvitationResponse, error) {
	vendorUUID, err := uuid.Parse(vendorID)
	if err != nil {
		return InvitationResponse{}, err
	}

	invitedByUUID, err := uuid.Parse(invitedByUserID)
	if err != nil {
		return InvitationResponse{}, err
	}

	roleUUID, err := uuid.Parse(dto.RoleID)
	if err != nil {
		return InvitationResponse{}, err
	}

	// Verify the inviter belongs to the vendor's organization
	inviterOrg, err := s.repo.GetOrganizationByUserID(ctx, invitedByUUID)
	if err != nil {
		return InvitationResponse{}, errors.New("only organization members can invite vendor users")
	}

	vendor, err := s.repo.GetVendorById(ctx, vendorUUID)
	if err != nil {
		return InvitationResponse{}, ErrVendorMismatch
	}

	if vendor.OrganizationID != inviterOrg.ID {
		return InvitationResponse{}, ErrVendorMismatch
	}

	token, err := generateToken()
	if err != nil {
		return InvitationResponse{}, err
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	invitation, err := s.repo.CreateVendorInvitation(ctx, repo.CreateVendorInvitationParams{
		VendorID:  vendorUUID,
		Email:     dto.Email,
		Token:     token,
		InvitedBy: invitedByUUID,
		RoleID:    roleUUID,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return InvitationResponse{}, err
	}
	return mapInvitationToResponse(invitation), nil
}

func (s *inviteService) AcceptInvitation(ctx context.Context, dto AcceptInviteDto) (*TokenResponse, error) {
	invitation, err := s.repo.GetVendorInvitationByToken(ctx, dto.Token)
	if err != nil {
		return nil, ErrInvitationNotFound
	}

	if invitation.Status != "pending" {
		if invitation.Status == "accepted" {
			return nil, ErrInvitationAlreadyUsed
		}
		return nil, ErrInvitationNotFound
	}

	if invitation.ExpiresAt.Time.Before(time.Now()) {
		_, _ = s.repo.UpdateVendorInvitationStatus(ctx, repo.UpdateVendorInvitationStatusParams{
			ID:     invitation.ID,
			Status: "expired",
		})
		return nil, ErrInvitationExpired
	}

	// Check if user already exists with this email
	existingUser, err := s.repo.GetUserByEmail(ctx, invitation.Email)
	if err == nil {
		if existingUser.OrganizationID.Valid {
			return nil, ErrVendorUserExists
		}
		if existingUser.VendorID.Valid {
			return nil, errors.New("a user with this email already belongs to a vendor")
		}
	}

	hashPass, err := utils.HashPassword(dto.Password)
	if err != nil {
		return nil, err
	}

	var user repo.User

	vendorUUID := pgtype.UUID{Bytes: invitation.VendorID, Valid: true}

	if existingUser.ID != uuid.Nil {
		// Link existing user to vendor
		user, err = s.repo.UpdateUserVendor(ctx, repo.UpdateUserVendorParams{
			ID:       existingUser.ID,
			VendorID: vendorUUID,
		})
		if err != nil {
			return nil, err
		}
	} else {
		// Create new user
		user, err = s.repo.CreateUser(ctx, repo.CreateUserParams{
			FirstName: dto.FirstName,
			LastName:  dto.LastName,
			Email:     invitation.Email,
			Password:  hashPass,
			Phone:     dto.Phone,
			RoleID:    invitation.RoleID,
		})
		if err != nil {
			return nil, err
		}
		// Link user to vendor
		user, err = s.repo.UpdateUserVendor(ctx, repo.UpdateUserVendorParams{
			ID:       user.ID,
			VendorID: vendorUUID,
		})
		if err != nil {
			return nil, err
		}
	}

	// Mark invitation as accepted
	_, err = s.repo.UpdateVendorInvitationStatus(ctx, repo.UpdateVendorInvitationStatusParams{
		ID:     invitation.ID,
		Status: "accepted",
	})
	if err != nil {
		return nil, err
	}

	// Generate JWT tokens
	tokens, err := jwt.GenerateTokenPair(user.ID.String(), user.RoleID.String(), s.jwtSecret, 15, 7)
	if err != nil {
		return nil, err
	}

	// Save refresh token
	_, err = s.repo.UpdateUserRefreshToken(ctx, repo.UpdateUserRefreshTokenParams{
		ID: user.ID,
		RefreshToken: pgtype.Text{
			String: tokens.RefreshToken,
			Valid:  true,
		},
		RefreshTokenExpiresAt: pgtype.Timestamptz{
			Time:  tokens.RefreshExp,
			Valid: true,
		},
	})
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *inviteService) GetInvitationsByVendor(ctx context.Context, vendorID, userID string) ([]InvitationResponse, error) {
	vendorUUID, err := uuid.Parse(vendorID)
	if err != nil {
		return nil, err
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	org, err := s.repo.GetOrganizationByUserID(ctx, userUUID)
	if err != nil {
		return nil, ErrAccessDenied
	}

	vendor, err := s.repo.GetVendorById(ctx, vendorUUID)
	if err != nil {
		return nil, err
	}

	if vendor.OrganizationID != org.ID {
		return nil, ErrAccessDenied
	}

	invitations, err := s.repo.GetVendorInvitationsByVendor(ctx, vendorUUID)
	if err != nil {
		return nil, err
	}

	response := make([]InvitationResponse, 0, len(invitations))
	for _, inv := range invitations {
		response = append(response, mapInvitationToResponse(inv))
	}
	return response, nil
}

func mapInvitationToResponse(inv repo.VendorInvitation) InvitationResponse {
	return InvitationResponse{
		ID:        inv.ID.String(),
		VendorID:  inv.VendorID.String(),
		Email:     inv.Email,
		RoleID:    inv.RoleID.String(),
		Status:    inv.Status,
		ExpiresAt: inv.ExpiresAt.Time.Format(time.RFC3339),
		CreatedAt: inv.CreatedAt.Time.Format(time.RFC3339),
	}
}
