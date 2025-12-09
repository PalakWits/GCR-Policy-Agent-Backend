package seller

import "time"

// SellerPendingCatalogSyncResponse defines the response body for the pending catalog sync sellers API
type SellerPendingCatalogSyncResponse struct {
	Domain       string       `json:"domain"`
	StatusFilter []string     `json:"status_filter"`
	Sellers      []SellerInfo `json:"sellers"`
	Page         PageInfo     `json:"page"`
}

// SellerInfo defines the structure for a seller's catalog sync information
type SellerInfo struct {
	SellerID      string     `json:"seller_id"`
	Status        string     `json:"status"`
	LastPullAt    *time.Time `json:"last_pull_at"`
	LastSuccessAt *time.Time `json:"last_success_at"`
	LastError     *string    `json:"last_error"`
}

// PageInfo defines the structure for pagination information
type PageInfo struct {
	Limit   int  `json:"limit"`
	Page    int  `json:"page"`
	HasMore bool `json:"has_more"`
}

// CatalogSyncStatusResponse defines the response body for the catalog sync status API
type SellerCatalogSyncStatusResponse struct {
	SellerID           string     `json:"seller_id"`
	Domain             string     `json:"domain"`
	Status             string     `json:"status"`
	LastPullAt         *time.Time `json:"last_pull_at"`
	LastSuccessAt      *time.Time `json:"last_success_at"`
	LastError          *string    `json:"last_error"`
	SyncVersion        int64      `json:"sync_version"`
	RegistryLastSeenAt time.Time  `json:"registry_last_seen_at"`
}

// SellerPermissionsUpdateRequest defines the structure for a single permission update
type SellerPermissionsUpdateRequest struct {
	SellerID       string     `json:"seller_id"`
	Domain         string     `json:"domain"`
	BapID          string     `json:"bap_id"`
	Decision       string     `json:"decision"`
	DecisionSource string     `json:"decision_source"`
	Reason         *string    `json:"reason"`
	ExpiresAt      *time.Time `json:"expires_at"`
}

// SellerPermissionsUpdateResponse defines the structure for a single permission update result
type SellerPermissionsUpdateResponse struct {
	SellerID string `json:"seller_id"`
	Domain   string `json:"domain"`
	BapID    string `json:"bap_id"`
	Decision string `json:"decision"`
	Stored   bool   `json:"stored"`
}

// SellerPermissionDetail provides detailed permission information for a single seller
type SellerPermissionDetail struct {
	SellerID       string     `json:"seller_id"`
	Domain         string     `json:"domain"`
	BapID          string     `json:"bap_id"`
	Decision       string     `json:"decision"`
	DecisionSource *string    `json:"decision_source,omitempty"`
	DecidedAt      *time.Time `json:"decided_at,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

// SellerRegistrySyncRequest defines the request body for the /v1/internal/registry-sync API
type SellerRegistrySyncRequest struct {
	Domains []string `json:"domains"`
}

// SellerDomainSyncSummary provides a summary of the sync operation for a single domain
type SellerDomainSyncSummary struct {
	Domain                 string `json:"domain"`
	NewSellers             int    `json:"new_sellers"`
	UpdatedSellers         int    `json:"updated_sellers"`
	DeactivatedSellers     int    `json:"deactivated_sellers"`
	TotalSellersInRegistry int    `json:"total_sellers_in_registry"`
}

// SellerRegistrySyncResponse defines the response body for the /v1/internal/registry-sync API
type SellerRegistrySyncResponse struct {
	Domains []SellerDomainSyncSummary `json:"domains"`
	RunAt   string                    `json:"run_at"`
}
