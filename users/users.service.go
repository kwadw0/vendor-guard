package users

import (
	"context"
	"vendor-guard/internal/repo"
	"vendor-guard/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserService interface {
	CreateUser(ctx context.Context, createUserDto CreateUserDto) (UserResponseDto, error)
}

type userService struct {
	repo *repo.Queries
}

func NewService(userRepo *repo.Queries) UserService {
	return &userService{repo: userRepo}
}

func (s *userService) CreateUser(ctx context.Context, createUserDto CreateUserDto) (UserResponseDto, error) {
	hashPass, err := utils.HashPassword(createUserDto.Password)
	if err != nil {
		return UserResponseDto{}, err
	}

	roleID, err := uuid.Parse(createUserDto.RoleID)
	if err != nil {
		return UserResponseDto{}, err
	}

	user, err := s.repo.CreateUser(ctx, repo.CreateUserParams{
		FirstName: createUserDto.FirstName,
		LastName:  createUserDto.LastName,
		Email:     createUserDto.Email,
		Password:  hashPass,
		Phone:     createUserDto.Phone,
		RoleID:    roleID,
		AvatarUrl: pgtype.Text{String: createUserDto.AvatarURL, Valid: createUserDto.AvatarURL != ""},
	})

	if err != nil {
		return UserResponseDto{}, err
	}
	return mapUserToResponse(user), nil
}

func mapUserToResponse(u repo.User) UserResponseDto {
	return UserResponseDto{
		ID:        u.ID.String(),
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Phone:     u.Phone,
		RoleID:    u.RoleID.String(),
		AvatarURL: u.AvatarUrl.String,
		CreatedAt: u.CreatedAt.Time.String(),
		UpdatedAt: u.UpdatedAt.Time.String(),
	}
}
