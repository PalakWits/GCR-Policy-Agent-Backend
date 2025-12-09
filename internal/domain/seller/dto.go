package seller

import (
	sellerPorts "adapter/internal/ports/seller"
	"adapter/internal/shared/crypto"

	"github.com/go-resty/resty/v2"
)

type SellerService struct {
	repo         sellerPorts.SellerRepository
	client       *resty.Client
	crypto       *crypto.ONDCCrypto
	domains      []string
	registryURL  string
	privateKey   string
	subscriberID string
	uniqueKeyID  string
}

type ONDCLookupRequest struct {
	Country string `json:"country"`
	Type    string `json:"type"`
	Domain  string `json:"domain"`
}

type Subscriber struct {
	SubscriberID  string `json:"subscriber_id"`
	UkID          string `json:"ukId"`
	BrID          string `json:"br_id"`
	Domain        string `json:"domain"`
	Country       string `json:"country"`
	City          string `json:"city"`
	SigningKey    string `json:"signing_public_key"`
	EncryptionKey string `json:"encr_public_key"`
	Status        string `json:"status"`
	ValidFrom     string `json:"valid_from"`
	ValidUntil    string `json:"valid_until"`
	Created       string `json:"created"`
	Updated       string `json:"updated"`
}

type ONDCLookupResponse []Subscriber
