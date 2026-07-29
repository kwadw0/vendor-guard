package vendors

type InviteVendorUserDto struct {
	Email  string `json:"email" validate:"required,email"`
	RoleID string `json:"role_id" validate:"required,uuid"`
}

type AcceptInviteDto struct {
	Token     string `json:"token" validate:"required"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Password  string `json:"password" validate:"required,min=8"`
	Phone     string `json:"phone" validate:"required"`
}

type InvitationResponse struct {
	ID        string `json:"id"`
	VendorID  string `json:"vendor_id"`
	Email     string `json:"email"`
	RoleID    string `json:"role_id"`
	Status    string `json:"status"`
	ExpiresAt string `json:"expires_at"`
	CreatedAt string `json:"created_at"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
