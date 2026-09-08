package users

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	appErrors "preuvio/internal/common"
	"preuvio/middleware"
	"preuvio/utils"
)

type Handler interface {
	CreateUser(w http.ResponseWriter, r *http.Request)
	GetUser(w http.ResponseWriter, r *http.Request)
	GetMe(w http.ResponseWriter, r *http.Request)
	GetAllUsers(w http.ResponseWriter, r *http.Request)
	UpdateUser(w http.ResponseWriter, r *http.Request)
	DeleteUser(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	service   UserService
	validator *validator.Validate
}

func NewHandler(service UserService, v *validator.Validate) Handler {
	return &handler{service: service, validator: v}
}

// CreateUser godoc
//
//	@Summary		Create a new user
//	@Description	Creates a new user in the system
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		CreateUserDto	true	"User creation data"
//	@Success		201		{object}	utils.SuccessResponse{data=users.UserResponseDto}
//	@Failure		400		{object}	utils.ErrorResponse
//	@Failure		500		{object}	utils.ErrorResponse
//	@Router			/api/users [post]
func (h *handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var dto CreateUserDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err)
		return
	}

	res, err := h.service.CreateUser(r.Context(), dto)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, "User created successfully", res)
}

// GetUser godoc
//
//	@Summary		Get a user by ID
//	@Description	Retrieve user details by ID
//	@Tags			users
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	utils.SuccessResponse{data=users.UserResponseDto}
//	@Failure		400	{object}	utils.ErrorResponse
//	@Failure		404	{object}	utils.ErrorResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/users/{id} [get]
func (h *handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.ErrorJSON(w, http.StatusBadRequest, errors.New("missing user id"))
		return
	}

	res, err := h.service.GetUser(r.Context(), id)
	if err != nil {
		if errors.Is(err, appErrors.ErrUserNotFound) {
			utils.ErrorJSON(w, http.StatusNotFound, err, "USER_NOT_FOUND")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, "User retrieved successfully", res)
}

// GetMe godoc
//
//	@Summary		Get current user
//	@Description	Returns the authenticated user from Bearer token
//	@Tags			users
//	@Produce		json
//	@Success		200	{object}	utils.SuccessResponse{data=users.UserResponseDto}	"Current user retrieved successfully"
//	@Failure		401	{object}	utils.ErrorResponse	"Unauthorized"
//	@Failure		404	{object}	utils.ErrorResponse	"User not found"
//	@Security		BearerAuth
//	@Router			/api/users/me [get]
func (h *handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		utils.ErrorJSON(w, http.StatusUnauthorized, errors.New("unauthorized"), "UNAUTHORIZED")
		return
	}
	res, err := h.service.GetUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, appErrors.ErrUserNotFound) {
			utils.ErrorJSON(w, http.StatusNotFound, err, "USER_NOT_FOUND")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Current user retrieved successfully", res)
}

// GetAllUsers godoc
//
//	@Summary		Get all users (paginated)
//	@Description	Retrieve a list of all users
//	@Tags			users
//	@Produce		json
//	@Param			page	query		int	false	"Page (1-based, default 1)"	Minimum(1)
//	@Param			limit	query		int	false	"Limit (default 20, max 50)"	Minimum(1)	Maximum(50)
//	@Success		200		{object}	utils.SuccessResponse{data=[]users.UserResponseDto,meta=utils.PaginationMeta}
//	@Failure		500		{object}	utils.ErrorResponse
//	@Router			/api/users [get]
func (h *handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	res, err := h.service.GetAllUsers(r.Context())
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, err)
		return
	}

	page, limit, _ := utils.ParsePagination(r.URL.Query(), utils.DefaultLimit, utils.MaxLimit)
	paged, meta := utils.PaginateSlice(res, page, limit)
	utils.WriteJSONWithMeta(w, http.StatusOK, "Users retrieved successfully", paged, meta)
}

// UpdateUser godoc
//
//	@Summary		Update a user
//	@Description	Update user details
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string			true	"User ID"
//	@Param			request	body		UpdateUserDto	true	"Update data"
//	@Success		200		{object}	utils.SuccessResponse{data=users.UserResponseDto}
//	@Failure		400		{object}	utils.ErrorResponse
//	@Failure		404		{object}	utils.ErrorResponse
//	@Failure		500		{object}	utils.ErrorResponse
//	@Router			/api/users/{id} [put]
func (h *handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.ErrorJSON(w, http.StatusBadRequest, errors.New("missing user id"))
		return
	}

	var dto UpdateUserDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err)
		return
	}

	res, err := h.service.UpdateUser(r.Context(), id, dto)
	if err != nil {
		if errors.Is(err, appErrors.ErrUserNotFound) {
			utils.ErrorJSON(w, http.StatusNotFound, err, "USER_NOT_FOUND")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, "User updated successfully", res)
}

// DeleteUser godoc
//
//	@Summary		Delete a user
//	@Description	Delete a user by ID
//	@Tags			users
//	@Param			id	path		string									true	"User ID"
//	@Success		200	{object}	utils.SuccessResponse{data=utils.EmptyData}	"User deleted successfully"
//	@Failure		400	{object}	utils.ErrorResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/users/{id} [delete]
func (h *handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.ErrorJSON(w, http.StatusBadRequest, errors.New("missing user id"))
		return
	}

	err := h.service.DeleteUser(r.Context(), id)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, "User deleted successfully", nil)
}
