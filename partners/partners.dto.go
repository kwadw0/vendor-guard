package partners

type CreatePartnerDto struct {
	Name  string `json:"name" validate:"required,min=2,max=255"`
	Email string `json:"email" validate:"omitempty,email,max=255"`
	Phone string `json:"phone" validate:"omitempty,max=255"`
}

type UpdatePartnerDto struct {
	ID    string
	Name  string `json:"name" validate:"required,min=2,max=255"`
	Email string `json:"email" validate:"omitempty,email,max=255"`
	Phone string `json:"phone" validate:"omitempty,max=255"`
}

type PartnerResponse struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`

	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`

	Status string `json:"status"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
