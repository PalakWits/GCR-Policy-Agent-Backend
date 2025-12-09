package buyer

type PermissionsRepository interface {
	UpsertBaps(baps map[string]Bap) error
	UpsertBapAccessPolicies(policies []BapAccessPolicy) error
	FindBapByID(bapID string) (*Bap, error)
	QueryBapAccessPolicies(bapID, domain string, sellerIDs []string) ([]BapAccessPolicy, error)
	GetBapPolicy(bapID string) (*BapAccessPolicy, error)
	CreatePermissionsJob(job *PermissionsJob) error
	UpdatePermissionsJobStatus(bapID, status string) error
	GetPermissionsJobStatus(bapID string) (*PermissionsJob, error)
}
