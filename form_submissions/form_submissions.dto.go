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

// Enriched response always returned by list endpoints (single query, no N+1)
type EnrichedSubmissionResponse struct {
	FormSubmissionResponse
	FormTitle    string `json:"form_title,omitempty"`
	FormStatus   string `json:"form_status,omitempty"`
	PartnerName  string `json:"partner_name,omitempty"`
	PartnerEmail string `json:"partner_email,omitempty"`
}

type SubmissionListQuery struct {
	Status    string `validate:"omitempty,oneof=pending approved rejected revision_required in_review"`
	FormID    string `validate:"omitempty,uuid"`
	PartnerID string `validate:"omitempty,uuid"`
	Q         string `validate:"omitempty,max=200"`
	Page      int    `validate:"omitempty,min=1"`
	Limit     int    `validate:"omitempty,min=1,max=50"`
	Sort      string `validate:"omitempty,oneof=submitted_at created_at"`
	Order     string `validate:"omitempty,oneof=asc desc"`
	DateFrom  string `validate:"omitempty"`
	DateTo    string `validate:"omitempty"`
}

type PaginatedSubmissionsResponse struct {
	Submissions []EnrichedSubmissionResponse `json:"submissions"`
	Total       int                          `json:"total"`
	Page        int                          `json:"page"`
	PageSize    int                          `json:"page_size"`
}

func formatOptionalTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
