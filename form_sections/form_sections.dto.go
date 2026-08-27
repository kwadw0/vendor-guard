package form_sections

type CreateSectionDto struct {
	Title       string `json:"title" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"omitempty,max=1000"`
	SortOrder   int    `json:"sort_order"`
}

type UpdateSectionDto struct {
	Title       string `json:"title" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"omitempty,max=1000"`
	SortOrder   int    `json:"sort_order"`
}

type SectionResponse struct {
	ID          string `json:"id"`
	FormID      string `json:"form_id,omitempty"`
	TemplateID  string `json:"template_id,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	SortOrder   int    `json:"sort_order"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
