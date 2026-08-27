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

type CloneFormResponse struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	TemplateID     string `json:"template_id,omitempty"`
	Title          string `json:"title"`
	Description    string `json:"description,omitempty"`
	Status         string `json:"status"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type TemplateSectionWithFields struct {
	Section SectionDetailResponse `json:"section"`
	Fields  []TemplateFieldDetail `json:"fields"`
}

type SectionDetailResponse struct {
	ID          string `json:"id"`
	TemplateID  string `json:"template_id,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	SortOrder   int    `json:"sort_order"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type TemplateFieldDetail struct {
	ID          string      `json:"id"`
	TemplateID  string      `json:"template_id"`
	SectionID   string      `json:"section_id"`
	FieldType   string      `json:"field_type"`
	Label       string      `json:"label"`
	Key         string      `json:"key"`
	Description string      `json:"description,omitempty"`
	Placeholder string      `json:"placeholder,omitempty"`
	IsRequired  bool        `json:"is_required"`
	SortOrder   int         `json:"sort_order"`
	Validation  interface{} `json:"validation"`
	Options     interface{} `json:"options"`
	CreatedAt   string      `json:"created_at"`
	UpdatedAt   string      `json:"updated_at"`
}

type TemplateDetailResponse struct {
	Template FormTemplateResponse        `json:"template"`
	Sections []TemplateSectionWithFields `json:"sections"`
}
