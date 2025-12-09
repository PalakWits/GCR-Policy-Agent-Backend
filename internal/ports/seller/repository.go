package seller

type SellerRepository interface {
	InsertSellers(sellers []Seller) error
	UpdateSellers(sellers []Seller) error
	GetAllSellers() ([]Seller, error)
	GetSellersByFilters(filters map[string]interface{}) ([]Seller, error)
	GetPendingSellers(domain, status string, limit, offset int) ([]SellerInfo, error)
	DeactivateSellers(sellerIDs []string, domain string) error
	UpsertCatalogState(state *SellerCatalogState) error
	GetSellerCatalogState(sellerID, domain string) (*SellerCatalogState, error)
}
