package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"time"

	"preuvio/auth/jwt"
	"preuvio/internal/repo"
	"preuvio/mail"
	"preuvio/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrEmailAlreadyExists  = errors.New("email already exists")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrInvalidVerifyToken  = errors.New("invalid or expired verification token")
	ErrInvalidResetToken   = errors.New("invalid or expired reset token")
)

type AuthService interface {
	Signup(ctx context.Context, dto SignupDto) (*TokenResponseDto, error)
	Login(ctx context.Context, dto LoginDto) (*TokenResponseDto, error)
	RefreshToken(ctx context.Context, dto RefreshTokenDto) (*TokenResponseDto, error)
	VerifyEmail(ctx context.Context, token string) error
	ForgotPassword(ctx context.Context, dto ForgotPasswordDto) error
	ResetPassword(ctx context.Context, dto ResetPasswordDto) error
}

type authService struct {
	repo      *repo.Queries
	jwtSecret string
	mailer    mail.Mailer
}

func NewService(repo *repo.Queries, secret string, mailers ...mail.Mailer) AuthService {
	s := &authService{repo: repo, jwtSecret: secret}
	if len(mailers) > 0 {
		s.mailer = mailers[0]
	}
	return s
}

func generateSecureToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
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

	// Verification email (best-effort: never fail signup on mail error)
	if token, err := generateSecureToken(32); err == nil {
		expiresAt := time.Now().Add(24 * time.Hour)
		if _, err := s.repo.CreateEmailVerificationToken(ctx, repo.CreateEmailVerificationTokenParams{
			UserID:    user.ID,
			Token:     token,
			ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
		}); err == nil {
			if s.mailer != nil {
				name := user.FirstName + " " + user.LastName
				if err := s.mailer.SendVerificationEmail(ctx, user.Email, name, token); err != nil {
					slog.Warn("failed to queue verification email", "email", user.Email, "error", err)
				}
			}
		} else {
			slog.Warn("failed to create verification token", "error", err)
		}
	}

	return s.generateAndSaveTokens(ctx, user.ID.String(), role.ID.String())
}

// VerifyEmail marks the user verified for a valid, unexpired token.
func (s *authService) VerifyEmail(ctx context.Context, token string) error {
	rec, err := s.repo.GetEmailVerificationByToken(ctx, token)
	if err != nil {
		return ErrInvalidVerifyToken
	}
	if rec.ExpiresAt.Valid && time.Now().After(rec.ExpiresAt.Time) {
		_ = s.repo.DeleteEmailVerificationByToken(ctx, token)
		return ErrInvalidVerifyToken
	}
	if _, err := s.repo.VerifyUserEmail(ctx, rec.UserID); err != nil {
		return err
	}
	_ = s.repo.DeleteEmailVerificationByUserID(ctx, rec.UserID)
	return nil
}

// ForgotPassword creates a reset token + queues the reset mail.
// Always returns nil for unknown emails to avoid account enumeration.
// The HTTP endpoint is wired; the function is ready even if frontend isn't.
func (s *authService) ForgotPassword(ctx context.Context, dto ForgotPasswordDto) error {
	user, err := s.repo.GetUserByEmail(ctx, dto.Email)
	if err != nil {
		return nil
	}
	token, err := generateSecureToken(32)
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(1 * time.Hour)
	if _, err := s.repo.CreatePasswordResetToken(ctx, repo.CreatePasswordResetTokenParams{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	}); err != nil {
		return err
	}
	if s.mailer != nil {
		name := user.FirstName + " " + user.LastName
		if err := s.mailer.SendPasswordResetEmail(ctx, user.Email, name, token); err != nil {
			slog.Warn("failed to queue password reset email", "email", user.Email, "error", err)
		}
	}
	return nil
}

// ResetPassword consumes a valid, unused, unexpired reset token.
func (s *authService) ResetPassword(ctx context.Context, dto ResetPasswordDto) error {
	rec, err := s.repo.GetPasswordResetByToken(ctx, dto.Token)
	if err != nil {
		return ErrInvalidResetToken
	}
	if rec.UsedAt.Valid {
		return ErrInvalidResetToken
	}
	if rec.ExpiresAt.Valid && time.Now().After(rec.ExpiresAt.Time) {
		return ErrInvalidResetToken
	}
	hash, err := utils.HashPassword(dto.NewPassword)
	if err != nil {
		return err
	}
	if _, err := s.repo.UpdateUserPassword(ctx, repo.UpdateUserPasswordParams{
		ID:       rec.UserID,
		Password: hash,
	}); err != nil {
		return err
	}
	if _, err := s.repo.MarkPasswordResetUsed(ctx, rec.ID); err != nil {
		return err
	}
	_ = s.repo.DeletePasswordResetByToken(ctx, dto.Token)
	_ = s.repo.RevokeRefreshToken(ctx, rec.UserID)
	return nil
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
