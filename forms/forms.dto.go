package forms

type CreateFormDto struct {
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

type FormSectionWithFields struct {
	Section SectionDetailResponse `json:"section"`
	Fields  []FormFieldDetail     `json:"fields"`
}

type SectionDetailResponse struct {
	ID          string `json:"id"`
	FormID      string `json:"form_id,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	SortOrder   int    `json:"sort_order"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type FormFieldDetail struct {
	ID          string      `json:"id"`
	FormID      string      `json:"form_id"`
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

type FormDetailResponse struct {
	Form     FormResponse            `json:"form"`
	Sections []FormSectionWithFields `json:"sections"`
}
