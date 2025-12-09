package broadcast

// SearchPayload defines the structure for the ONDC /search request
type SearchPayload struct {
	Context *Context `json:"context"`
	Message *Message `json:"message"`
}

// BroadcastRequest defines the request body for the /v1/permissions/broadcast API
type BroadcastRequest struct {
	SearchPayload   *SearchPayload `json:"search_payload"`
	SellerIDs       []string       `json:"seller_ids,omitempty"`
	IncludeNoPolicy bool           `json:"include_no_policy,omitempty"`
}

// Context defines the context for the /search API
type Context struct {
	Domain        string `json:"domain"`
	Action        string `json:"action"`
	Country       string `json:"country"`
	City          string `json:"city"`
	CoreVersion   string `json:"core_version"`
	BapID         string `json:"bap_id"`
	BapURI        string `json:"bap_uri"`
	TransactionID string `json:"transaction_id"`
	MessageID     string `json:"message_id"`
	Timestamp     string `json:"timestamp"`
	TTL           string `json:"ttl"`
}

// Message defines the message for the /search API
type Message struct {
	Intent *Intent `json:"intent"`
}

// Intent defines the intent for the /search API
type Intent struct {
	Category    *Category    `json:"category"`
	Fulfillment *Fulfillment `json:"fulfillment"`
	Payment     *Payment     `json:"payment"`
	Tags        []*Tag       `json:"tags"`
}

// Category defines the category for the /search API
type Category struct {
	ID string `json:"id"`
}

// Fulfillment defines the fulfillment for the /search API
type Fulfillment struct {
	Type string `json:"type"`
}

// Payment defines the payment for the /search API
type Payment struct {
	BuyerAppFinderFeeType   string `json:"@ondc/org/buyer_app_finder_fee_type"`
	BuyerAppFinderFeeAmount string `json:"@ondc/org/buyer_app_finder_fee_amount"`
}

// Tag defines the tags for the /search API
type Tag struct {
	Code string  `json:"code"`
	List []*List `json:"list"`
}

// List defines the list for the /search API
type List struct {
	Code  string `json:"code"`
	Value string `json:"value"`
}

// AckResponse defines the response body for the /search API
type AckResponse struct {
	Context *Context    `json:"context"`
	Message *AckMessage `json:"message"`
}

// AckMessage defines the message for the /search API
type AckMessage struct {
	Ack *Ack `json:"ack"`
}

// Ack defines the ack for the /search API
type Ack struct {
	Status string `json:"status"`
}

// NackResponse defines the response body for the /search API
type NackResponse struct {
	Context *Context    `json:"context"`
	Message *AckMessage `json:"message"`
	Error   *Error      `json:"error"`
}

// Error defines the error for the /search API
type Error struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
}
