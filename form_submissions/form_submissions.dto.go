package form_submissions

import "time"

type CreateFormSubmissionDto struct {
	Responses map[string]interface{} `json:"responses" validate:"required"`
}

type ReviewFormSubmissionDto struct {
	Status string `json:"status" validate:"required,oneof=approved rejected revision_required"`
}

type FormSubmissionResponse struct {
	ID          string                 `json:"id"`
	FormID      string                 `json:"form_id"`
	PartnerID   string                 `json:"partner_id"`
	SubmittedBy string                 `json:"submitted_by,omitempty"`
	Status      string                 `json:"status"`
	Responses   map[string]interface{} `json:"responses"`
	SubmittedAt string                 `json:"submitted_at,omitempty"`
	ReviewedAt  string                 `json:"reviewed_at,omitempty"`
	ReviewedBy  string                 `json:"reviewed_by,omitempty"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

type PaginatedSubmissionsResponse struct {
	Submissions []FormSubmissionResponse `json:"submissions"`
	Total       int                      `json:"total"`
	Page        int                      `json:"page"`
	PageSize    int                      `json:"page_size"`
}

func formatOptionalTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
