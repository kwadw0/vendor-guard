package form_templates

type CreateFormTemplateDto struct {
	Title       string `json:"title" validate:"required,min=2,max=255"`
	Description string `json:"description" validate:"omitempty,max=1000"`
	Category    string `json:"category" validate:"omitempty,max=100"`
}

type UpdateFormTemplateDto struct {
	Title       string `json:"title" validate:"required,min=2,max=255"`
	Description string `json:"description" validate:"omitempty,max=1000"`
	Category    string `json:"category" validate:"omitempty,max=100"`
	IsActive    bool   `json:"is_active"`
}

type FormTemplateResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type CloneTemplateDto struct {
	Title string `json:"title" validate:"required,min=2,max=255"`
}
