package auth

import (
	"context"
	"errors"

	"vendor-guard/auth/jwt"
	"vendor-guard/internal/repo"
	"vendor-guard/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrEmailAlreadyExists  = errors.New("email already exists")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
)

type AuthService interface {
	Signup(ctx context.Context, dto SignupDto) (*TokenResponseDto, error)
	Login(ctx context.Context, dto LoginDto) (*TokenResponseDto, error)
	RefreshToken(ctx context.Context, dto RefreshTokenDto) (*TokenResponseDto, error)
}

type authService struct {
	repo      *repo.Queries
	jwtSecret string
}

func NewService(repo *repo.Queries, secret string) AuthService {
	return &authService{repo: repo, jwtSecret: secret}
}

func (s *authService) Signup(ctx context.Context, dto SignupDto) (*TokenResponseDto, error) {
	_, err := s.repo.GetUserByEmail(ctx, dto.Email)
	if err == nil {
		return nil, ErrEmailAlreadyExists
	}

	hashPass, err := utils.HashPassword(dto.Password)
	if err != nil {
		return nil, err
	}

	role, err := s.repo.GetRoleByName(ctx, "member")
	if err != nil {
		return nil, err
	}

	user, err := s.repo.CreateUser(ctx, repo.CreateUserParams{
		FirstName: dto.FirstName,
		LastName:  dto.LastName,
		Email:     dto.Email,
		Password:  hashPass,
		Phone:     dto.Phone,
		RoleID:    role.ID,
	})
	if err != nil {
		return nil, err
	}

	return s.generateAndSaveTokens(ctx, user.ID.String(), role.ID.String())
}

func (s *authService) Login(ctx context.Context, dto LoginDto) (*TokenResponseDto, error) {
	user, err := s.repo.GetUserByEmail(ctx, dto.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := utils.VerifyPassword(user.Password, dto.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateAndSaveTokens(ctx, user.ID.String(), user.RoleID.String())
}

func (s *authService) RefreshToken(ctx context.Context, dto RefreshTokenDto) (*TokenResponseDto, error) {
	userID, err := jwt.ValidateRefreshToken(dto.RefreshToken, s.jwtSecret)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	user, err := s.repo.GetUserByRefreshToken(ctx, pgtype.Text{String: dto.RefreshToken, Valid: true})
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	if user.ID.String() != userID {
		return nil, ErrInvalidRefreshToken
	}

	return s.generateAndSaveTokens(ctx, user.ID.String(), user.RoleID.String())
}

func (s *authService) generateAndSaveTokens(ctx context.Context, userID, roleID string) (*TokenResponseDto, error) {
	tokens, err := jwt.GenerateTokenPair(userID, roleID, s.jwtSecret, 15, 7)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	_, err = s.repo.UpdateUserRefreshToken(ctx, repo.UpdateUserRefreshTokenParams{
		ID: id,
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

	return &TokenResponseDto{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}
