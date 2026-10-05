package auth

import (
	"errors"
	"net/http"

	"preuvio/utils"

	"github.com/go-playground/validator/v10"
)

type Handler interface {
	Signup(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	RefreshToken(w http.ResponseWriter, r *http.Request)
	VerifyEmail(w http.ResponseWriter, r *http.Request)
	ForgotPassword(w http.ResponseWriter, r *http.Request)
	ResetPassword(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	service   AuthService
	validator *validator.Validate
}

func NewHandler(service AuthService, v *validator.Validate) Handler {
	return &handler{service: service, validator: v}
}

// Signup godoc
//
//	@Summary		Register a new user
//	@Description	Creates a new user and returns access and refresh tokens
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		SignupDto								true	"User signup data"
//	@Success		201		{object}	utils.SuccessResponse{data=auth.TokenResponseDto}	"Signup successful"
//	@Failure		400		{object}	utils.ErrorResponse
//	@Failure		409		{object}	utils.ErrorResponse
//	@Failure		500		{object}	utils.ErrorResponse
//	@Router			/api/auth/signup [post]
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
//
//	@Summary		User login
//	@Description	Authenticates a user and returns access and refresh tokens
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		LoginDto								true	"User login credentials"
//	@Success		200		{object}	utils.SuccessResponse{data=auth.TokenResponseDto}	"Login successful"
//	@Failure		400		{object}	utils.ErrorResponse
//	@Failure		401		{object}	utils.ErrorResponse
//	@Failure		500		{object}	utils.ErrorResponse
//	@Router			/api/auth/login [post]
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
//
//	@Summary		Refresh access token
//	@Description	Generates a new pair of access and refresh tokens using a valid refresh token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RefreshTokenDto							true	"Refresh token"
//	@Success		200		{object}	utils.SuccessResponse{data=auth.TokenResponseDto}	"Token refreshed successfully"
//	@Failure		400		{object}	utils.ErrorResponse
//	@Failure		401		{object}	utils.ErrorResponse
//	@Failure		500		{object}	utils.ErrorResponse
//	@Router			/api/auth/refresh [post]
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

// VerifyEmail godoc
//
//	@Summary		Verify email address
//	@Description	Verifies a user's email using the token sent at signup
//	@Tags			auth
//	@Produce		json
//	@Param			token	query		string	true	"Verification token"
//	@Success		200		{object}	utils.SuccessResponse	"Email verified successfully"
//	@Failure		400		{object}	utils.ErrorResponse
//	@Router			/api/auth/verify-email [get]
func (h *handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		utils.ErrorJSON(w, http.StatusBadRequest, errors.New("missing token"), "VALIDATION_ERROR")
		return
	}
	if err := h.service.VerifyEmail(r.Context(), token); err != nil {
		if errors.Is(err, ErrInvalidVerifyToken) {
			utils.ErrorJSON(w, http.StatusBadRequest, err, "INVALID_TOKEN")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Email verified successfully", nil)
}

// ForgotPassword godoc
//
//	@Summary		Request password reset
//	@Description	Creates a reset token and emails a reset link (always 200 to avoid enumeration)
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		ForgotPasswordDto	true	"Email address"
//	@Success		200		{object}	utils.SuccessResponse	"Reset link sent if account exists"
//	@Failure		400		{object}	utils.ErrorResponse
//	@Router			/api/auth/forgot-password [post]
func (h *handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var dto ForgotPasswordDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "BAD_REQUEST")
		return
	}
	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}
	if err := h.service.ForgotPassword(r.Context(), dto); err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "If an account exists, a reset link has been sent", nil)
}

// ResetPassword godoc
//
//	@Summary		Reset password with token
//	@Description	Consumes a valid reset token and sets a new password
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		ResetPasswordDto	true	"Token and new password"
//	@Success		200		{object}	utils.SuccessResponse	"Password reset successfully"
//	@Failure		400		{object}	utils.ErrorResponse
//	@Router			/api/auth/reset-password [post]
func (h *handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var dto ResetPasswordDto
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "BAD_REQUEST")
		return
	}
	if err := h.validator.Struct(dto); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err, "VALIDATION_ERROR")
		return
	}
	if err := h.service.ResetPassword(r.Context(), dto); err != nil {
		if errors.Is(err, ErrInvalidResetToken) {
			utils.ErrorJSON(w, http.StatusBadRequest, err, "INVALID_TOKEN")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, "Password reset successfully", nil)
}
