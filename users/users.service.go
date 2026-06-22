package users

import (
	"context"
	appErrors "vendor-guard/internal/common"
	"vendor-guard/internal/repo"
	"vendor-guard/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserService interface {
	CreateUser(ctx context.Context, dto CreateUserDto) (UserResponseDto, error)
	GetUser(ctx context.Context, userID string) (UserResponseDto, error)
	GetAllUsers(ctx context.Context) ([]UserResponseDto, error)
	UpdateUser(ctx context.Context, userID string, dto UpdateUserDto) (UserResponseDto, error)
	DeleteUser(ctx context.Context, userID string) error
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

func (s *userService) GetUser(ctx context.Context, userID string) (UserResponseDto, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return UserResponseDto{}, err
	}

	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return UserResponseDto{}, err
	}

	return mapUserToResponse(user), nil
}

func (s *userService) GetAllUsers(
	ctx context.Context,
) ([]UserResponseDto, error) {

	users, err := s.repo.ListUsers(ctx, repo.ListUsersParams{
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		return nil, err
	}

	response := make([]UserResponseDto, 0, len(users))

	for _, user := range users {
		response = append(response, mapUserToResponse(user))
	}

	return response, nil
}

func (s *userService) UpdateUser(
	ctx context.Context,
	userID string,
	dto UpdateUserDto,
) (UserResponseDto, error) {

	id, err := uuid.Parse(userID)
	if err != nil {
		return UserResponseDto{}, err
	}

	existingUser, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return UserResponseDto{}, appErrors.ErrUserNotFound
	}

	roleID := existingUser.RoleID

	if dto.RoleID != nil {
		roleID, err = uuid.Parse(*dto.RoleID)
		if err != nil {
			return UserResponseDto{}, err
		}
	}

	firstName := existingUser.FirstName
	if dto.FirstName != nil {
		firstName = *dto.FirstName
	}

	lastName := existingUser.LastName
	if dto.LastName != nil {
		lastName = *dto.LastName
	}

	email := existingUser.Email
	if dto.Email != nil {
		email = *dto.Email
	}

	phone := existingUser.Phone
	if dto.Phone != nil {
		phone = *dto.Phone
	}

	avatar := existingUser.AvatarUrl.String
	if dto.AvatarURL != nil {
		avatar = *dto.AvatarURL
	}

	updatedUser, err := s.repo.UpdateUser(ctx, repo.UpdateUserParams{
		ID:        id,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Phone:     phone,
		RoleID:    roleID,
		AvatarUrl: pgtype.Text{
			String: avatar,
			Valid:  avatar != "",
		},
	})

	if err != nil {
		return UserResponseDto{}, err
	}

	return mapUserToResponse(updatedUser), nil
}

func (s *userService) DeleteUser(
	ctx context.Context,
	userID string,
) error {

	id, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	return s.repo.DeleteUser(ctx, id)

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
