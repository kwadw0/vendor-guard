package partners

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"preuvio/auth/jwt"
	"preuvio/internal/repo"
	"preuvio/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrInvitationNotFound    = errors.New("invitation not found or expired")
	ErrInvitationExpired     = errors.New("invitation has expired")
	ErrInvitationAlreadyUsed = errors.New("invitation already accepted")
	ErrPartnerMismatch       = errors.New("partner does not belong to your organization")
	ErrPartnerUserExists     = errors.New("a user with this email already belongs to an organization")
)

type InviteService interface {
	InviteUser(ctx context.Context, partnerID, invitedByUserID string, dto InvitePartnerUserDto) (InvitationResponse, error)
	AcceptInvitation(ctx context.Context, dto AcceptInviteDto) (*TokenResponse, error)
	GetInvitationsByPartner(ctx context.Context, partnerID, userID string) ([]InvitationResponse, error)
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

func (s *inviteService) InviteUser(ctx context.Context, partnerID, invitedByUserID string, dto InvitePartnerUserDto) (InvitationResponse, error) {
	partnerUUID, err := uuid.Parse(partnerID)
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

	// Verify the inviter belongs to the partner's organization
	inviterOrg, err := s.repo.GetOrganizationByUserID(ctx, invitedByUUID)
	if err != nil {
		return InvitationResponse{}, errors.New("only organization members can invite partner users")
	}

	partner, err := s.repo.GetPartnerById(ctx, partnerUUID)
	if err != nil {
		return InvitationResponse{}, ErrPartnerMismatch
	}

	if partner.OrganizationID != inviterOrg.ID {
		return InvitationResponse{}, ErrPartnerMismatch
	}

	token, err := generateToken()
	if err != nil {
		return InvitationResponse{}, err
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	invitation, err := s.repo.CreatePartnerInvitation(ctx, repo.CreatePartnerInvitationParams{
		PartnerID: partnerUUID,
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
	invitation, err := s.repo.GetPartnerInvitationByToken(ctx, dto.Token)
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
		_, _ = s.repo.UpdatePartnerInvitationStatus(ctx, repo.UpdatePartnerInvitationStatusParams{
			ID:     invitation.ID,
			Status: "expired",
		})
		return nil, ErrInvitationExpired
	}

	// Check if user already exists with this email
	existingUser, err := s.repo.GetUserByEmail(ctx, invitation.Email)
	if err == nil {
		if existingUser.OrganizationID.Valid {
			return nil, ErrPartnerUserExists
		}
		if existingUser.PartnerID.Valid {
			return nil, errors.New("a user with this email already belongs to a partner")
		}
	}

	hashPass, err := utils.HashPassword(dto.Password)
	if err != nil {
		return nil, err
	}

	var user repo.User

	partnerUUID := pgtype.UUID{Bytes: invitation.PartnerID, Valid: true}

	if existingUser.ID != uuid.Nil {
		// Link existing user to partner
		user, err = s.repo.UpdateUserPartner(ctx, repo.UpdateUserPartnerParams{
			ID:        existingUser.ID,
			PartnerID: partnerUUID,
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
		// Link user to partner
		user, err = s.repo.UpdateUserPartner(ctx, repo.UpdateUserPartnerParams{
			ID:        user.ID,
			PartnerID: partnerUUID,
		})
		if err != nil {
			return nil, err
		}
	}

	// Mark invitation as accepted
	_, err = s.repo.UpdatePartnerInvitationStatus(ctx, repo.UpdatePartnerInvitationStatusParams{
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

func (s *inviteService) GetInvitationsByPartner(ctx context.Context, partnerID, userID string) ([]InvitationResponse, error) {
	partnerUUID, err := uuid.Parse(partnerID)
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

	partner, err := s.repo.GetPartnerById(ctx, partnerUUID)
	if err != nil {
		return nil, err
	}

	if partner.OrganizationID != org.ID {
		return nil, ErrAccessDenied
	}

	invitations, err := s.repo.GetPartnerInvitationsByPartner(ctx, partnerUUID)
	if err != nil {
		return nil, err
	}

	response := make([]InvitationResponse, 0, len(invitations))
	for _, inv := range invitations {
		response = append(response, mapInvitationToResponse(inv))
	}
	return response, nil
}

func mapInvitationToResponse(inv repo.PartnerInvitation) InvitationResponse {
	return InvitationResponse{
		ID:        inv.ID.String(),
		PartnerID: inv.PartnerID.String(),
		Email:     inv.Email,
		RoleID:    inv.RoleID.String(),
		Status:    inv.Status,
		ExpiresAt: inv.ExpiresAt.Time.Format(time.RFC3339),
		CreatedAt: inv.CreatedAt.Time.Format(time.RFC3339),
	}
}
