package organizations

import (
	"time"

	"github.com/google/uuid"
)

type CreateOrganizationDto struct {
	Name                string `json:"name" validate:"required"`
	Description         string `json:"description"`
	WebsiteURL          string `json:"website_url" validate:"omitempty,url"`
	Industry            string `json:"industry" validate:"required"`
	TeamSize            string `json:"team_size" validate:"required"`
	PrimaryCustomerType string `json:"primary_customer_type" validate:"required,oneof=b2b b2c both"`
	OwnerRole           string `json:"owner_role" validate:"required"`
}

type UpdateOrganizationDto struct {
	Name                string `json:"name" validate:"required"`
	Description         string `json:"description"`
	WebsiteURL          string `json:"website_url" validate:"omitempty,url"`
	Industry            string `json:"industry" validate:"required"`
	TeamSize            string `json:"team_size" validate:"required"`
	PrimaryCustomerType string `json:"primary_customer_type" validate:"required,oneof=b2b b2c both"`
	OwnerRole           string `json:"owner_role" validate:"required"`
}

type OrganizationResponseDto struct {
	ID                  uuid.UUID `json:"id"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	WebsiteURL          string    `json:"website_url"`
	Industry            string    `json:"industry"`
	TeamSize            string    `json:"team_size"`
	PrimaryCustomerType string    `json:"primary_customer_type"`
	OwnerRole           string    `json:"owner_role"`
	IsActive            bool      `json:"is_active"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
