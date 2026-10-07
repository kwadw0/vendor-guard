package form_assignments

import (
	"context"
	"errors"
	"time"

	"preuvio/internal/repo"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var ErrAssignmentNotFound = errors.New("assignment not found")
var ErrAccessDenied = errors.New("access denied")
var ErrFormNotFound = errors.New("form not found")
var ErrFormNotActive = errors.New("form is not active")

type AssignmentService interface {
	AssignForm(ctx context.Context, formID string, dto AssignFormDto, userID string) (AssignResultResponse, error)
	GetAssignmentsByForm(ctx context.Context, formID string, status string, userID string) ([]AssignmentResponse, error)
	GetAssignmentsByPartner(ctx context.Context, partnerID string, status string, userID string) ([]AssignmentResponse, error)
	GetMyAssignments(ctx context.Context, status string, userID string) ([]AssignmentResponse, error)
	RevokeAssignment(ctx context.Context, formID string, partnerID string, userID string) error
}

type assignmentService struct {
	repo *repo.Queries
}

func NewService(queries *repo.Queries) AssignmentService {
	return &assignmentService{repo: queries}
}

// AssignForm creates live assignments for the given partners.
// Idempotent: pairs already assigned are reported in Skipped, never duplicated
// (partial UNIQUE index + ON CONFLICT DO NOTHING; race losers re-read the row).
func (s *assignmentService) AssignForm(ctx context.Context, formID string, dto AssignFormDto, userID string) (AssignResultResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return AssignResultResponse{}, err
	}
	formUUID, err := uuid.Parse(formID)
	if err != nil {
		return AssignResultResponse{}, err
	}

	org, err := s.repo.GetOrganizationByUserID(ctx, userUUID)
	if err != nil {
		return AssignResultResponse{}, ErrAccessDenied
	}

	form, err := s.repo.GetFormByID(ctx, formUUID)
	if err != nil {
		return AssignResultResponse{}, ErrFormNotFound
	}
	if form.OrganizationID != org.ID {
		return AssignResultResponse{}, ErrAccessDenied
	}
	if form.Status != "active" {
		return AssignResultResponse{}, ErrFormNotActive
	}

	var dueAt pgtype.Timestamptz
	if dto.DueAt != "" {
		t, err := time.Parse(time.RFC3339, dto.DueAt)
		if err != nil {
			if t2, err2 := time.Parse("2006-01-02", dto.DueAt); err2 != nil {
				return AssignResultResponse{}, err
			} else {
				t = t2
			}
		}
		dueAt = pgtype.Timestamptz{Time: t, Valid: true}
	}

	result := AssignResultResponse{
		Assigned: []AssignmentResponse{},
		Skipped:  []SkippedAssignment{},
	}
	for _, pid := range dto.PartnerIDs {
		partnerUUID, err := uuid.Parse(pid)
		if err != nil {
			result.Skipped = append(result.Skipped, SkippedAssignment{PartnerID: pid, Reason: "invalid partner id"})
			continue
		}
		partner, err := s.repo.GetPartnerById(ctx, partnerUUID)
		if err != nil {
			result.Skipped = append(result.Skipped, SkippedAssignment{PartnerID: pid, Reason: "partner not found"})
			continue
		}
		if partner.OrganizationID != org.ID {
			result.Skipped = append(result.Skipped, SkippedAssignment{PartnerID: pid, Reason: "partner belongs to another organization"})
			continue
		}
		if !partner.Status.Valid || string(partner.Status.PartnerStatus) != "active" {
			result.Skipped = append(result.Skipped, SkippedAssignment{PartnerID: pid, Reason: "partner is not active (status: " + string(partner.Status.PartnerStatus) + ")"})
			continue
		}

		row, err := s.repo.AssignFormToPartner(ctx, repo.AssignFormToPartnerParams{
			FormID:    formUUID,
			PartnerID: partnerUUID,
			AssignedBy: pgtype.UUID{
				Bytes: userUUID,
				Valid: true,
			},
			DueAt: dueAt,
		})
		if err != nil {
			// ON CONFLICT DO NOTHING returns no row: pair already assigned.
			live, lErr := s.repo.GetLiveAssignment(ctx, repo.GetLiveAssignmentParams{
				FormID:    formUUID,
				PartnerID: partnerUUID,
			})
			if lErr != nil {
				result.Skipped = append(result.Skipped, SkippedAssignment{PartnerID: pid, Reason: "failed to assign"})
				continue
			}
			row = live
		}
		result.Assigned = append(result.Assigned, mapAssignmentToResponse(row, partner.Name, partner.Email, "", "", 0))
	}
	return result, nil
}

func (s *assignmentService) GetAssignmentsByForm(ctx context.Context, formID string, status string, userID string) ([]AssignmentResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	formUUID, err := uuid.Parse(formID)
	if err != nil {
		return nil, err
	}
	org, err := s.repo.GetOrganizationByUserID(ctx, userUUID)
	if err != nil {
		return nil, ErrAccessDenied
	}
	form, err := s.repo.GetFormByID(ctx, formUUID)
	if err != nil {
		return nil, ErrFormNotFound
	}
	if form.OrganizationID != org.ID {
		return nil, ErrAccessDenied
	}

	rows, err := s.repo.GetAssignmentsByFormID(ctx, repo.GetAssignmentsByFormIDParams{
		FormID:       formUUID,
		StatusFilter: pgtype.Text{String: status, Valid: status != ""},
	})
	if err != nil {
		return nil, err
	}
	response := make([]AssignmentResponse, 0, len(rows))
	for _, r := range rows {
		response = append(response, AssignmentResponse{
			ID:               r.ID.String(),
			FormID:           r.FormID.String(),
			PartnerID:        r.PartnerID.String(),
			Status:           presentedStatus(r.Status, r.DueAt),
			DueAt:            formatOptionalTime(r.DueAt.Time),
			SubmittedAt:      formatOptionalTime(r.SubmittedAt.Time),
			CreatedAt:        r.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt:        r.UpdatedAt.Time.Format(time.RFC3339),
			PartnerName:      r.PartnerName,
			PartnerEmail:     r.PartnerEmail,
			SubmissionsCount: int(r.SubmissionsCount),
		})
	}
	return response, nil
}

func (s *assignmentService) GetAssignmentsByPartner(ctx context.Context, partnerID string, status string, userID string) ([]AssignmentResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	partnerUUID, err := uuid.Parse(partnerID)
	if err != nil {
		return nil, err
	}
	org, err := s.repo.GetOrganizationByUserID(ctx, userUUID)
	if err != nil {
		return nil, ErrAccessDenied
	}
	partner, err := s.repo.GetPartnerById(ctx, partnerUUID)
	if err != nil {
		return nil, ErrAssignmentNotFound
	}
	if partner.OrganizationID != org.ID {
		return nil, ErrAccessDenied
	}
	return s.listByPartner(ctx, partnerUUID, status)
}

// GetMyAssignments is the vendor work queue: forced to the caller's own
// partner, so a partner can never list another vendor's work.
func (s *assignmentService) GetMyAssignments(ctx context.Context, status string, userID string) ([]AssignmentResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	partner, err := s.repo.GetPartnerByUserID(ctx, userUUID)
	if err != nil {
		return nil, ErrAccessDenied
	}
	return s.listByPartner(ctx, partner.ID, status)
}

func (s *assignmentService) listByPartner(ctx context.Context, partnerID uuid.UUID, status string) ([]AssignmentResponse, error) {
	rows, err := s.repo.GetAssignmentsByPartnerID(ctx, repo.GetAssignmentsByPartnerIDParams{
		PartnerID:    partnerID,
		StatusFilter: pgtype.Text{String: status, Valid: status != ""},
	})
	if err != nil {
		return nil, err
	}
	response := make([]AssignmentResponse, 0, len(rows))
	for _, r := range rows {
		response = append(response, AssignmentResponse{
			ID:          r.ID.String(),
			FormID:      r.FormID.String(),
			PartnerID:   r.PartnerID.String(),
			Status:      presentedStatus(r.Status, r.DueAt),
			DueAt:       formatOptionalTime(r.DueAt.Time),
			SubmittedAt: formatOptionalTime(r.SubmittedAt.Time),
			CreatedAt:   r.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt:   r.UpdatedAt.Time.Format(time.RFC3339),
			FormTitle:   r.FormTitle,
			FormStatus:  r.FormStatus,
		})
	}
	return response, nil
}

func (s *assignmentService) RevokeAssignment(ctx context.Context, formID string, partnerID string, userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	formUUID, err := uuid.Parse(formID)
	if err != nil {
		return err
	}
	partnerUUID, err := uuid.Parse(partnerID)
	if err != nil {
		return err
	}
	org, err := s.repo.GetOrganizationByUserID(ctx, userUUID)
	if err != nil {
		return ErrAccessDenied
	}
	form, err := s.repo.GetFormByID(ctx, formUUID)
	if err != nil {
		return ErrFormNotFound
	}
	if form.OrganizationID != org.ID {
		return ErrAccessDenied
	}
	live, err := s.repo.GetLiveAssignment(ctx, repo.GetLiveAssignmentParams{
		FormID:    formUUID,
		PartnerID: partnerUUID,
	})
	if err != nil {
		return ErrAssignmentNotFound
	}
	_, err = s.repo.UpdateAssignmentStatus(ctx, repo.UpdateAssignmentStatusParams{
		ID:          live.ID,
		Status:      "revoked",
		SubmittedAt: pgtype.Timestamptz{Valid: false},
	})
	return err
}

// presentedStatus computes overdue at read time: an assigned row past due_at
// reads as overdue without any cron or row flip.
func presentedStatus(stored string, dueAt pgtype.Timestamptz) string {
	if stored == "assigned" && dueAt.Valid && time.Now().After(dueAt.Time) {
		return "overdue"
	}
	return stored
}

func mapAssignmentToResponse(a repo.FormAssignment, partnerName, partnerEmail, formTitle, formStatus string, count int) AssignmentResponse {
	return AssignmentResponse{
		ID:               a.ID.String(),
		FormID:           a.FormID.String(),
		PartnerID:        a.PartnerID.String(),
		Status:           presentedStatus(a.Status, a.DueAt),
		DueAt:            formatOptionalTime(a.DueAt.Time),
		SubmittedAt:      formatOptionalTime(a.SubmittedAt.Time),
		CreatedAt:        a.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:        a.UpdatedAt.Time.Format(time.RFC3339),
		PartnerName:      partnerName,
		PartnerEmail:     partnerEmail,
		FormTitle:        formTitle,
		FormStatus:       formStatus,
		SubmissionsCount: count,
	}
}

func formatOptionalTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
