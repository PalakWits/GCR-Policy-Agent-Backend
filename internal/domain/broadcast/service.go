package broadcast

import (
	"adapter/internal/config"
	"adapter/internal/ports/broadcast"
	"adapter/internal/ports/buyer"
	sellerPorts "adapter/internal/ports/seller"
	"adapter/internal/shared/log"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"gorm.io/gorm"
	url_pkg "net/url"
	"path"
	"strings"
	"time"
)

type BroadcastService struct {
	buyerRepo  buyer.PermissionsRepository
	sellerRepo sellerPorts.SellerRepository
	httpClient *resty.Client
	config     *config.Config
}

func NewBroadcastService(buyerRepo buyer.PermissionsRepository, sellerRepo sellerPorts.SellerRepository, cfg *config.Config) *BroadcastService {
	return &BroadcastService{
		buyerRepo:  buyerRepo,
		sellerRepo: sellerRepo,
		httpClient: resty.New().SetTimeout(60 * time.Second),
		config:     cfg,
	}
}

func (s *BroadcastService) BroadcastPermissions(req broadcast.BroadcastRequest) error {
	ctx := context.Background() // Initialize context for logging

	if req.SearchPayload == nil || req.SearchPayload.Context == nil {
		log.Errorf(ctx, nil, "search_payload and its context are required in broadcast request")
		return fmt.Errorf("search_payload and its context are required")
	}
	bapID := req.SearchPayload.Context.BapID
	if bapID == "" {
		log.Errorf(ctx, nil, "bap_id is required in search_payload context")
		return fmt.Errorf("bap_id is required")
	}

	// Check if BAP exists, create if not.
	_, err := s.buyerRepo.FindBapByID(bapID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			now := time.Now()
			newBap := buyer.Bap{BapID: bapID, FirstSeenAt: now, LastSeenAt: now}
			if err := s.buyerRepo.UpsertBaps(map[string]buyer.Bap{bapID: newBap}); err != nil {
				log.Errorf(ctx, err, "Failed to create new BAP with id %s", bapID)
				return fmt.Errorf("failed to create new BAP with id %s: %w", bapID, err)
			}
		} else {
			log.Errorf(ctx, err, "Database error when trying to find BAP by ID %s", bapID)
			return err
		}
	}

	policy, err := s.buyerRepo.GetBapPolicy(bapID)
	if err != nil {
		// This check is tricky now because GetBapPolicy might return gorm.ErrRecordNotFound
		// which is a normal case here. We only care about actual errors.
		if err != gorm.ErrRecordNotFound {
			log.Errorf(ctx, err, "Database error when trying to get BAP policy for bap_id %s", bapID)
			return err
		}
	}

	if policy != nil {
		log.Warnf(ctx, "Policy already exists for bap_id: %s. Not initiating broadcast.", bapID)
		return fmt.Errorf("policy already exists for bap_id: %s", bapID)
	}

	job := &buyer.PermissionsJob{
		BapID:  bapID,
		Status: "INITIATED",
	}

	if err := s.buyerRepo.CreatePermissionsJob(job); err != nil {
		log.Errorf(ctx, err, "Failed to create permissions job for bap_id %s", bapID)
		return err
	}

	log.Infof(ctx, "Initiating broadcast for bap_id %s", bapID)
	go s.startBroadcast(req)

	return nil
}

func (s *BroadcastService) startBroadcast(req broadcast.BroadcastRequest) {
	ctx := context.Background()
	domain := req.SearchPayload.Context.Domain
	sellerIDs := req.SellerIDs

	filters := map[string]interface{}{
		"domain": domain,
		"active": true,
	}

	if len(sellerIDs) > 0 {
		log.Infof(ctx, "Constructing filter for specific sellers: %v", sellerIDs)
		filters["seller_id"] = sellerIDs
	} else {
		city := req.SearchPayload.Context.City
		log.Infof(ctx, "Constructing filter for all sellers in domain '%s' and city '%s' (or wildcard)", domain, city)
		filters["city"] = city
	}

	sellers, err := s.sellerRepo.GetSellersByFilters(filters)
	if err != nil {
		log.Errorf(ctx, err, "Failed to fetch sellers for broadcast")
		s.updateJobStatus(req.SearchPayload.Context.BapID, "FAILED")
		return
	}

	if len(sellers) == 0 {
		log.Warnf(ctx, "No sellers found for broadcast criteria. Marking job as COMPLETED.")
		s.updateJobStatus(req.SearchPayload.Context.BapID, "COMPLETED")
		return
	}

	log.Infof(ctx, "Starting broadcast to %d sellers for bap_id %s", len(sellers), req.SearchPayload.Context.BapID)
	done := make(chan bool, len(sellers))

	for _, sel := range sellers {
		go s.sendSearchRequest(sel, req, done)
	}

	for i := 0; i < len(sellers); i++ {
		<-done
	}

	log.Infof(ctx, "Broadcast finished for bap_id %s. All sellers have responded.", req.SearchPayload.Context.BapID)
	s.updateJobStatus(req.SearchPayload.Context.BapID, "COMPLETED")
}

func (s *BroadcastService) updateJobStatus(bapID, status string) {
	err := s.buyerRepo.UpdatePermissionsJobStatus(bapID, status)
	if err != nil {
		// Handle error
	}
}

func (s *BroadcastService) GetBroadcastStatus(bapID string) (*buyer.PermissionsJob, error) {
	return s.buyerRepo.GetPermissionsJobStatus(bapID)
}

func (s *BroadcastService) sendSearchRequest(seller sellerPorts.Seller, req broadcast.BroadcastRequest, done chan bool) {
	defer func() { done <- true }()
	ctx := context.Background() // Create a new context for logging

	var policy *buyer.BapAccessPolicy
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	if s.config.MockSellerResponse {
		log.Infof(ctx, "MOCK_SELLER_RESPONSE is true. Returning mock ACK for seller %s", seller.SellerID)
		// Construct a mock ACK response to simulate success
		reason := "ALLOWED for testing"
		policy = &buyer.BapAccessPolicy{
			SellerID:  seller.SellerID,
			Domain:    req.SearchPayload.Context.Domain,
			BapID:     req.SearchPayload.Context.BapID,
			Decision:  sellerPorts.DecisionAllowed,
			DecidedAt: now,
			ExpiresAt: &expiresAt,
			Reason:    &reason,
		}
	} else {

		searchReqPayload := map[string]interface{}{
			"context": req.SearchPayload.Context,
			"message": req.SearchPayload.Message,
		}

		// Ensure the URL has a scheme.
		url := seller.SubscriberURL
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			url = "https://" + url
		}
		// It seems some subscriber URLs might already contain a path.
		// We need to correctly append /search. Let's use url.Parse.
		parsedURL, err := url_pkg.Parse(url)
		if err != nil {
			log.Errorf(ctx, err, "Failed to parse seller SubscriberURL: %s", seller.SubscriberURL)
			return
		}
		parsedURL.Path = path.Join(parsedURL.Path, "search")
		finalURL := parsedURL.String()

		log.Infof(ctx, "Sending /search request to seller %s at %s", seller.SellerID, finalURL)

		resp, err := s.httpClient.R().
			SetBody(searchReqPayload).
			Post(finalURL)

		if err != nil {
			log.Errorf(ctx, err, "Failed to send /search request to seller %s", seller.SellerID)
			return
		}

		log.Infof(ctx, "Received response from seller %s: Status %d, Body: %s", seller.SellerID, resp.StatusCode(), string(resp.Body()))

		if resp.IsSuccess() {
			var ackResponse broadcast.AckResponse
			if err := json.Unmarshal(resp.Body(), &ackResponse); err == nil && ackResponse.Message != nil && ackResponse.Message.Ack != nil && ackResponse.Message.Ack.Status == "ACK" {
				log.Infof(ctx, "Received ACK from seller %s for bap_id %s. Creating GRANTED policy.", seller.SellerID, req.SearchPayload.Context.BapID)
				policy = &buyer.BapAccessPolicy{
					SellerID:  seller.SellerID,
					Domain:    req.SearchPayload.Context.Domain,
					BapID:     req.SearchPayload.Context.BapID,
					Decision:  sellerPorts.DecisionAllowed,
					DecidedAt: now,
					ExpiresAt: &expiresAt,
				}
			} else {

				var nackResponse broadcast.NackResponse
				if err := json.Unmarshal(resp.Body(), &nackResponse); err == nil && nackResponse.Message != nil && nackResponse.Message.Ack != nil && nackResponse.Message.Ack.Status == "NACK" {
					log.Infof(ctx, "Received NACK from seller %s for bap_id %s. Creating DENIED policy.", seller.SellerID, req.SearchPayload.Context.BapID)
					reason := "NACK received from seller"
					if nackResponse.Error != nil {
						reason = nackResponse.Error.Message
					}
					policy = &buyer.BapAccessPolicy{
						SellerID:  seller.SellerID,
						Domain:    req.SearchPayload.Context.Domain,
						BapID:     req.SearchPayload.Context.BapID,
						Decision:  sellerPorts.DecisionDenied,
						Reason:    &reason,
						DecidedAt: now,
						ExpiresAt: &expiresAt,
					}
				} else {

					log.Warnf(ctx, "Received success status from seller %s, but could not decode ACK/NACK from body: %s", seller.SellerID, string(resp.Body()))
				}
			}
		} else {
			log.Errorf(ctx, nil, "Received non-success status (%d) from seller %s", resp.StatusCode(), seller.SellerID)
			reason := fmt.Sprintf("Received non-success status %d from seller. Body: %s", resp.StatusCode(), string(resp.Body()))
			policy = &buyer.BapAccessPolicy{
				SellerID:  seller.SellerID,
				Domain:    req.SearchPayload.Context.Domain,
				BapID:     req.SearchPayload.Context.BapID,
				Decision:  sellerPorts.DecisionErrorOccurred,
				Reason:    &reason,
				DecidedAt: now,
				ExpiresAt: &expiresAt,
			}
		}
	}

	if policy != nil {
		if err := s.buyerRepo.UpsertBapAccessPolicies([]buyer.BapAccessPolicy{*policy}); err != nil {
			log.Errorf(ctx, err, "Failed to upsert BapAccessPolicy for seller %s and bap_id %s", seller.SellerID, req.SearchPayload.Context.BapID)
		} else {
			log.Infof(ctx, "Successfully upserted BapAccessPolicy for seller %s and bap_id %s with decision %s", seller.SellerID, req.SearchPayload.Context.BapID, policy.Decision)
		}
	}
}
