package form_assignments

type AssignFormDto struct {
	PartnerIDs []string `json:"partner_ids" validate:"required,min=1,max=100,dive,uuid"`
	DueAt      string   `json:"due_at" validate:"omitempty"`
}

type AssignmentResponse struct {
	ID              string `json:"id"`
	FormID          string `json:"form_id"`
	PartnerID       string `json:"partner_id"`
	Status          string `json:"status"`
	DueAt           string `json:"due_at,omitempty"`
	SubmittedAt     string `json:"submitted_at,omitempty"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
	PartnerName     string `json:"partner_name,omitempty"`
	PartnerEmail    string `json:"partner_email,omitempty"`
	FormTitle       string `json:"form_title,omitempty"`
	FormStatus      string `json:"form_status,omitempty"`
	SubmissionsCount int   `json:"submissions_count,omitempty"`
}

type AssignResultResponse struct {
	Assigned []AssignmentResponse `json:"assigned"`
	Skipped  []SkippedAssignment  `json:"skipped"`
}

type SkippedAssignment struct {
	PartnerID string `json:"partner_id"`
	Reason    string `json:"reason"`
}
