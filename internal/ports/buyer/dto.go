package buyer

import (
	"adapter/internal/ports/seller"
)

// BapPermissionsQueryRequest defines the request body for the /v1/permissions/query API
type BapPermissionsQueryRequest struct {
	BapID           string   `json:"bap_id"`
	Domain          string   `json:"domain"`
	RegistryEnv     string   `json:"registry_env"`
	SellerIDs       []string `json:"seller_ids"`
	IncludeNoPolicy bool     `json:"include_no_policy"`
}

// BapPermissionsQueryResponse defines the response body for the /v1/permissions/query API
type BapPermissionsQueryResponse struct {
	BapStatus   string                          `json:"bap_status"`
	Domain      string                          `json:"domain"`
	RegistryEnv string                          `json:"registry_env"`
	Permissions []seller.SellerPermissionDetail `json:"permissions"`
}
