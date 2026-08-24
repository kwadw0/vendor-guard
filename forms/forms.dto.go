package forms

type CreateFormDto struct {
	TemplateID  string `json:"template_id" validate:"omitempty,uuid"`
	Title       string `json:"title" validate:"required,min=2,max=255"`
	Description string `json:"description" validate:"omitempty,max=1000"`
}

type UpdateFormDto struct {
	Title       string `json:"title" validate:"required,min=2,max=255"`
	Description string `json:"description" validate:"omitempty,max=1000"`
	Status      string `json:"status" validate:"omitempty,oneof=draft active archived"`
}

type FormResponse struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	TemplateID     string `json:"template_id,omitempty"`
	Title          string `json:"title"`
	Description    string `json:"description,omitempty"`
	Status         string `json:"status"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}
