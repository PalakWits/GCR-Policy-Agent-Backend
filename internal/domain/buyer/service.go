package buyer

import (
	buyerPorts "adapter/internal/ports/buyer"
	sellerPorts "adapter/internal/ports/seller"

	"gorm.io/gorm"
	"time"
)

type BuyerService struct {
	repo buyerPorts.PermissionsRepository
}

func NewBuyerService(repo buyerPorts.PermissionsRepository) *BuyerService {
	return &BuyerService{repo: repo}
}

func (s *BuyerService) UpdateBapAccessPermissions(updates []sellerPorts.SellerPermissionsUpdateRequest) ([]sellerPorts.SellerPermissionsUpdateResponse, error) {
	var results []sellerPorts.SellerPermissionsUpdateResponse
	var policiesToUpsert []buyerPorts.BapAccessPolicy
	bapsToUpsert := make(map[string]buyerPorts.Bap)

	for _, update := range updates {
		// Prepare BapAccessPolicy for upsert
		policy := buyerPorts.BapAccessPolicy{
			SellerID:       update.SellerID,
			Domain:         update.Domain,
			RegistryEnv:    update.RegistryEnv,
			BapID:          update.BapID,
			Decision:       sellerPorts.AccessDecision(update.Decision),
			DecisionSource: sellerPorts.DecisionSource(update.DecisionSource),
			DecidedAt:      time.Now(),
			ExpiresAt:      update.ExpiresAt,
			Reason:         update.Reason,
		}
		policiesToUpsert = append(policiesToUpsert, policy)

		// Collect unique BAPs to ensure they exist in the `baps` table
		if _, exists := bapsToUpsert[update.BapID]; !exists {
			bapsToUpsert[update.BapID] = buyerPorts.Bap{BapID: update.BapID}
		}

		results = append(results, sellerPorts.SellerPermissionsUpdateResponse{
			SellerID:    update.SellerID,
			Domain:      update.Domain,
			RegistryEnv: update.RegistryEnv,
			BapID:       update.BapID,
			Decision:    update.Decision,
			Stored:      false, // Will be set to true after successful DB operation
		})
	}

	// Upsert BAPs first to satisfy foreign key constraints
	if err := s.repo.UpsertBaps(bapsToUpsert); err != nil {
		return results, err
	}

	// Upsert the access policies
	if err := s.repo.UpsertBapAccessPolicies(policiesToUpsert); err != nil {
		return results, err
	}

	// Mark all as stored if we reach here without an error
	for i := range results {
		results[i].Stored = true
	}

	return results, nil
}

func (s *BuyerService) QueryBapAccessPermissions(req buyerPorts.BapPermissionsQueryRequest) (*buyerPorts.BapPermissionsQueryResponse, error) {
	bapStatus := ""
	bap, err := s.repo.FindBapByID(req.BapID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			bapStatus = "NEW_BAP"
			// Create the BAP
			bapsToUpsert := map[string]buyerPorts.Bap{req.BapID: {BapID: req.BapID}}
			if err := s.repo.UpsertBaps(bapsToUpsert); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		bapStatus = "EXISTING_BAP"
		// Update last_seen_at
		bap.LastSeenAt = time.Now()
		bapsToUpsert := map[string]buyerPorts.Bap{req.BapID: *bap}
		if err := s.repo.UpsertBaps(bapsToUpsert); err != nil {
			return nil, err
		}
	}

	policies, err := s.repo.QueryBapAccessPolicies(req.BapID, req.Domain, req.RegistryEnv, req.SellerIDs)
	if err != nil {
		return nil, err
	}

	policyMap := make(map[string]buyerPorts.BapAccessPolicy)
	for _, p := range policies {
		policyMap[p.SellerID] = p
	}

	var permissions []sellerPorts.SellerPermissionDetail
	for _, sellerID := range req.SellerIDs {
		if policy, ok := policyMap[sellerID]; ok {
			permissions = append(permissions, sellerPorts.SellerPermissionDetail{
				SellerID:       policy.SellerID,
				Domain:         policy.Domain,
				RegistryEnv:    policy.RegistryEnv,
				BapID:          policy.BapID,
				Decision:       string(policy.Decision),
				DecisionSource: (*string)(&policy.DecisionSource),
				DecidedAt:      &policy.DecidedAt,
				ExpiresAt:      policy.ExpiresAt,
			})
		} else if req.IncludeNoPolicy {
			permissions = append(permissions, sellerPorts.SellerPermissionDetail{
				SellerID:    sellerID,
				Domain:      req.Domain,
				RegistryEnv: req.RegistryEnv,
				BapID:       req.BapID,
				Decision:    "NO_POLICY",
			})
		}
	}

	return &buyerPorts.BapPermissionsQueryResponse{
		BapStatus:   bapStatus,
		Domain:      req.Domain,
		RegistryEnv: req.RegistryEnv,
		Permissions: permissions,
	}, nil
}
