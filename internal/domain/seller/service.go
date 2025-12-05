package seller

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"adapter/internal/config"
	sellerPorts "adapter/internal/ports/seller"
	"adapter/internal/shared/crypto"
	"adapter/internal/shared/log"

	"github.com/go-resty/resty/v2"
	"gorm.io/gorm"
)

func NewSellerService(repo sellerPorts.SellerRepository, cfg *config.Config) *SellerService {
	client := resty.New()
	client.SetTimeout(30 * time.Second)
	client.SetRetryCount(3)
	client.SetRetryWaitTime(5 * time.Second)

	return &SellerService{
		repo:         repo,
		client:       client,
		crypto:       crypto.NewONDCCrypto(),
		domains:      cfg.Domains,
		registryURL:  cfg.RegistryURL,
		privateKey:   cfg.PrivateKey,
		subscriberID: cfg.SubscriberID,
		uniqueKeyID:  cfg.UniqueKeyID,
		registryEnv:  cfg.RegistryEnv,
	}
}

func (s *SellerService) GetPendingCatalogSyncSellers(domain, registryEnv, status string, limit, page, offset int) (*sellerPorts.SellerPendingCatalogSyncResponse, error) {
	sellers, err := s.repo.GetPendingSellers(domain, registryEnv, status, limit, offset)
	if err != nil {
		return nil, err
	}

	hasMore := len(sellers) > limit
	if hasMore {
		sellers = sellers[:limit] // Trim the extra record fetched for hasMore check
	}

	var statusFilter []string
	if status == "" {
		statusFilter = []string{"NOT_SYNCED", "FAILED"}
	} else {
		statusFilter = strings.Split(status, ",")
	}

	return &sellerPorts.SellerPendingCatalogSyncResponse{
		Domain:       domain,
		RegistryEnv:  registryEnv,
		StatusFilter: statusFilter,
		Sellers:      sellers,
		Page: sellerPorts.PageInfo{
			Limit:   limit,
			Page:    page,
			HasMore: hasMore,
		},
	}, nil
}

func (s *SellerService) GetSyncStatus(sellerID, domain, registryEnv string) (*sellerPorts.SellerCatalogSyncStatusResponse, error) {
	seller, err := s.repo.GetSellerByID(sellerID, domain, registryEnv)
	if err != nil {
		return nil, err // Could be gorm.ErrRecordNotFound
	}

	state, err := s.repo.GetSellerCatalogState(sellerID, domain, registryEnv)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// If state is not found, we still return a response but with default/empty values for state fields
	if err == gorm.ErrRecordNotFound {
		return &sellerPorts.SellerCatalogSyncStatusResponse{
			SellerID:           seller.SellerID,
			Domain:             seller.Domain,
			RegistryEnv:        seller.RegistryEnv,
			Status:             string(sellerPorts.CatalogStatusNotSynced), // Default status
			RegistryLastSeenAt: seller.LastSeenInReg,
		}, nil
	}

	return &sellerPorts.SellerCatalogSyncStatusResponse{
		SellerID:           state.SellerID,
		Domain:             state.Domain,
		RegistryEnv:        state.RegistryEnv,
		Status:             string(state.Status),
		LastPullAt:         state.LastPullAt,
		LastSuccessAt:      state.LastSuccessAt,
		LastError:          state.LastError,
		SyncVersion:        state.SyncVersion,
		RegistryLastSeenAt: seller.LastSeenInReg,
	}, nil
}

func (s *SellerService) SyncRegistry(req sellerPorts.SellerRegistrySyncRequest) (*sellerPorts.SellerRegistrySyncResponse, error) {
	runAt := time.Now()
	response := &sellerPorts.SellerRegistrySyncResponse{
		RegistryEnv: req.RegistryEnv,
		RunAt:       runAt.Format(time.RFC3339),
		Domains:     []sellerPorts.SellerDomainSyncSummary{},
	}

	for _, domain := range req.Domains {
		summary := sellerPorts.SellerDomainSyncSummary{Domain: domain}

		registrySellers, err := s.FetchSellersFromRegistry(domain)
		if err != nil {
			log.Error(context.Background(), err, fmt.Sprintf("Failed to fetch sellers from registry for domain %s", domain))
			continue
		}
		summary.TotalSellersInRegistry = len(registrySellers)

		dbSellers, err := s.repo.GetSellersByDomainAndRegistry(domain, req.RegistryEnv)
		if err != nil {
			log.Error(context.Background(), err, fmt.Sprintf("Failed to fetch sellers from DB for domain %s", domain))
			continue
		}

		registrySellerMap := make(map[string]sellerPorts.Seller)
		now := time.Now()
		for _, sub := range registrySellers {
			validFrom, _ := time.Parse(time.RFC3339, sub.ValidFrom)
			validUntil, _ := time.Parse(time.RFC3339, sub.ValidUntil)
			raw, _ := json.Marshal(sub)

			seller := sellerPorts.Seller{
				SellerID: sub.SubscriberID, Domain: sub.Domain, RegistryEnv: req.RegistryEnv,
				Status: sub.Status, Type: "BPP", SubscriberURL: sub.SubscriberID,
				Country: sub.Country, City: sub.City, ValidFrom: validFrom, ValidUntil: validUntil,
				Active: true, LastSeenInReg: now, RegistryRaw: string(raw),
			}
			registrySellerMap[seller.SellerID] = seller
		}

		dbSellerMap := make(map[string]sellerPorts.Seller)
		for _, seller := range dbSellers {
			dbSellerMap[seller.SellerID] = seller
		}

		var sellersToInsert []sellerPorts.Seller
		var sellersToUpdate []sellerPorts.Seller
		var removedSellerIDs []string

		for id, seller := range registrySellerMap {
			if _, exists := dbSellerMap[id]; !exists {
				sellersToInsert = append(sellersToInsert, seller)
			} else {
				sellersToUpdate = append(sellersToUpdate, seller)
			}
		}

		for id := range dbSellerMap {
			if _, exists := registrySellerMap[id]; !exists {
				removedSellerIDs = append(removedSellerIDs, id)
			}
		}

		summary.NewSellers = len(sellersToInsert)
		summary.UpdatedSellers = len(sellersToUpdate)
		summary.DeactivatedSellers = len(removedSellerIDs)

		if len(sellersToInsert) > 0 {
			if err := s.repo.InsertSellers(sellersToInsert); err != nil {
				log.Error(context.Background(), err, "Failed to insert new sellers")
			}
		}
		if len(sellersToUpdate) > 0 {
			if err := s.repo.UpdateSellers(sellersToUpdate); err != nil {
				log.Error(context.Background(), err, "Failed to update existing sellers")
			}
		}
		for _, seller := range sellersToInsert {
			state := &sellerPorts.SellerCatalogState{
				SellerID: seller.SellerID, Domain: domain, RegistryEnv: req.RegistryEnv,
				Status: sellerPorts.CatalogStatusNotSynced,
			}
			if err := s.repo.UpsertCatalogState(state); err != nil {
				log.Error(context.Background(), err, "Failed to insert catalog state")
			}
		}
		if len(removedSellerIDs) > 0 {
			if err := s.repo.DeactivateSellers(removedSellerIDs, domain, req.RegistryEnv); err != nil {
				log.Error(context.Background(), err, "Failed to deactivate sellers")
			}
		}
		response.Domains = append(response.Domains, summary)
	}
	return response, nil
}

func (s *SellerService) FetchSellersFromRegistry(domain string) (ONDCLookupResponse, error) {
	reqBody := ONDCLookupRequest{Country: "IND", Type: "BPP", Domain: domain}
	authHeader, err := s.generateAuthHeader(reqBody)
	if err != nil {
		return nil, err
	}
	var response ONDCLookupResponse
	resp, err := s.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", authHeader).
		SetBody(reqBody).
		SetResult(&response).
		Post(s.registryURL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode(), resp.String())
	}
	return response, nil
}

func (s *SellerService) generateAuthHeader(body ONDCLookupRequest) (string, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	if s.privateKey == "" {
		return "", fmt.Errorf("PRIVATE_KEY not configured")
	}
	currentTime := int(time.Now().Unix())
	ttl := 30
	signature, err := s.crypto.SignRequest(s.privateKey, payload, currentTime, ttl)
	if err != nil {
		return "", err
	}
	authHeader := fmt.Sprintf(
		`Signature keyId="%s|%s|ed25519",algorithm="ed25519",created="%d",expires="%d",headers="(created) (expires) digest",signature="%s"`,
		s.subscriberID, s.uniqueKeyID, currentTime, currentTime+ttl, signature,
	)
	return authHeader, nil
}
