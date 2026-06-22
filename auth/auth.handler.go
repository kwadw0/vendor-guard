package auth

import (
	"errors"
	"net/http"

	"vendor-guard/utils"

	"github.com/go-playground/validator/v10"
)

type Handler interface {
	Signup(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	RefreshToken(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	service   AuthService
	validator *validator.Validate
}

func NewHandler(service AuthService, v *validator.Validate) Handler {
	return &handler{service: service, validator: v}
}

// Signup godoc
// @Summary Register a new user
// @Description Creates a new user and returns access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body SignupDto true "User signup data"
// @Success 201 {object} utils.SuccessResponse[TokenResponseDto] "Signup successful"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/auth/signup [post]
func (h *handler) Signup(w http.ResponseWriter, r *http.Request) {
	var dto SignupDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "BAD_REQUEST")
		return
	}

	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}

	res, err := h.service.Signup(r.Context(), dto)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			utils.ErrorJSON(w, http.StatusConflict, err, "EMAIL_EXISTS")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, "User created successfully", res)
}

// Login godoc
// @Summary User login
// @Description Authenticates a user and returns access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginDto true "User login credentials"
// @Success 200 {object} utils.SuccessResponse[TokenResponseDto] "Login successful"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/auth/login [post]
func (h *handler) Login(w http.ResponseWriter, r *http.Request) {
	var dto LoginDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "BAD_REQUEST")
		return
	}

	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}

	res, err := h.service.Login(r.Context(), dto)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			utils.ErrorJSON(w, http.StatusUnauthorized, err, "INVALID_CREDENTIALS")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}

	utils.WriteJSON(w, http.StatusOK, "Login successful", res)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Generates a new pair of access and refresh tokens using a valid refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenDto true "Refresh token"
// @Success 200 {object} utils.SuccessResponse[TokenResponseDto] "Token refreshed successfully"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/auth/refresh [post]
func (h *handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var dto RefreshTokenDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "BAD_REQUEST")
		return
	}

	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}

	res, err := h.service.RefreshToken(r.Context(), dto)
	if err != nil {
		if errors.Is(err, ErrInvalidRefreshToken) {
			utils.ErrorJSON(w, http.StatusUnauthorized, err, "INVALID_REFRESH_TOKEN")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}

	utils.WriteJSON(w, http.StatusOK, "Token refreshed successfully", res)
}
