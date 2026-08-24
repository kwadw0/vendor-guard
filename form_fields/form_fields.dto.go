package form_fields

type CreateFormFieldDto struct {
	SectionID   string      `json:"section_id" validate:"required,uuid"`
	FieldType   string      `json:"field_type" validate:"required,oneof=text textarea number email phone date checkbox radio select multiselect file richtext"`
	Label       string      `json:"label" validate:"required,min=1,max=255"`
	Key         string      `json:"key" validate:"required,min=1,max=100,alphanum"`
	Description string      `json:"description" validate:"omitempty,max=1000"`
	Placeholder string      `json:"placeholder" validate:"omitempty,max=255"`
	IsRequired  bool        `json:"is_required"`
	SortOrder   int         `json:"sort_order"`
	Validation  interface{} `json:"validation"`
	Options     interface{} `json:"options"`
}

type UpdateFormFieldDto struct {
	FieldType   string      `json:"field_type" validate:"required,oneof=text textarea number email phone date checkbox radio select multiselect file richtext"`
	Label       string      `json:"label" validate:"required,min=1,max=255"`
	Key         string      `json:"key" validate:"required,min=1,max=100,alphanum"`
	Description string      `json:"description" validate:"omitempty,max=1000"`
	Placeholder string      `json:"placeholder" validate:"omitempty,max=255"`
	IsRequired  bool        `json:"is_required"`
	SortOrder   int         `json:"sort_order"`
	Validation  interface{} `json:"validation"`
	Options     interface{} `json:"options"`
}

type FormFieldResponse struct {
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
