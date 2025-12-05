package buyer

import (
	"time"

	seller "adapter/internal/ports/seller"
)

type Bap struct {
	BapID       string    `gorm:"primaryKey;column:bap_id;type:text"`
	FirstSeenAt time.Time `gorm:"column:first_seen_at;type:timestamptz;autoCreateTime"`
	LastSeenAt  time.Time `gorm:"column:last_seen_at;type:timestamptz;autoUpdateTime"`
}

func (Bap) TableName() string {
	return "baps"
}

type BapAccessPolicy struct {
	SellerID       string                `gorm:"primaryKey;column:seller_id;type:text"`
	Domain         string                `gorm:"primaryKey;column:domain;type:text"`
	RegistryEnv    string                `gorm:"primaryKey;column:registry_env;type:text"`
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
