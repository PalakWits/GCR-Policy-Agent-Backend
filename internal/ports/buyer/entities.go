package buyer

import (
	"time"

	"adapter/internal/ports/seller"
	"github.com/google/uuid"
)

// Bap defines the structure for a BAP (Buyer App)
type Bap struct {
	BapID       string    `json:"bap_id" gorm:"primary_key"`
	FirstSeenAt time.Time `json:"first_seen_at" gorm:"autoCreateTime"`
	LastSeenAt  time.Time `json:"last_seen_at" gorm:"autoUpdateTime"`
}

func (Bap) TableName() string {
	return "baps"
}

type BapAccessPolicy struct {
	SellerID       string                `gorm:"primaryKey;column:seller_id;type:text"`
	Domain         string                `gorm:"primaryKey;column:domain;type:text"`
	BapID          string                `gorm:"primaryKey;column:bap_id;type:text"`
	Decision       seller.AccessDecision `gorm:"column:decision;type:text"`
	DecisionSource seller.DecisionSource `gorm:"column:decision_source;type:text"`
	DecidedAt      time.Time             `gorm:"column:decided_at;type:timestamptz"`
	ExpiresAt      *time.Time            `gorm:"column:expires_at;type:timestamptz"`
	Reason         *string               `gorm:"column:reason;type:text"`
	UpdatedAt      time.Time             `gorm:"column:updated_at;type:timestamptz;autoUpdateTime"`
}

func (BapAccessPolicy) TableName() string {
	return "bap_access_policy"
}

type PermissionsJob struct {
	ID        uuid.UUID `json:"job_id" gorm:"type:uuid;default:gen_random_uuid();primary_key"`
	BapID     string    `json:"bap_id" gorm:"not null"`
	Status    string    `json:"status" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
