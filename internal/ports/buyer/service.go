package buyer

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BuyerRepository struct {
	db *gorm.DB
}

func NewBuyerRepository(db *gorm.DB) *BuyerRepository {
	return &BuyerRepository{db: db}
}

func (r *BuyerRepository) UpsertBaps(baps map[string]Bap) error {
	var bapList []Bap
	for _, b := range baps {
		bapList = append(bapList, b)
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "bap_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"last_seen_at"}),
	}).Create(&bapList).Error
}

func (r *BuyerRepository) UpsertBapAccessPolicies(policies []BapAccessPolicy) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "seller_id"}, {Name: "domain"}, {Name: "bap_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"decision", "decision_source", "decided_at", "expires_at", "reason", "updated_at"}),
	}).Create(&policies).Error
}
func (r *BuyerRepository) FindBapByID(bapID string) (*Bap, error) {
	var bap Bap
	if err := r.db.First(&bap, "bap_id = ?", bapID).Error; err != nil {
		return nil, err
	}
	return &bap, nil
}

func (r *BuyerRepository) QueryBapAccessPolicies(bapID, domain string, sellerIDs []string) ([]BapAccessPolicy, error) {
	var policies []BapAccessPolicy
	if err := r.db.Where("bap_id = ? AND domain = ? AND seller_id IN ?", bapID, domain, sellerIDs).Find(&policies).Error; err != nil {
		return nil, err
	}
	return policies, nil
}

func (r *BuyerRepository) GetBapPolicy(bapID string) (*BapAccessPolicy, error) {
	var policy BapAccessPolicy
	if err := r.db.Where("bap_id = ?", bapID).First(&policy).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &policy, nil
}

func (r *BuyerRepository) CreatePermissionsJob(job *PermissionsJob) error {
	return r.db.Create(job).Error
}

func (r *BuyerRepository) UpdatePermissionsJobStatus(jobID uuid.UUID, status string) error {
	return r.db.Model(&PermissionsJob{}).Where("id = ?", jobID).Update("status", status).Error
}

func (r *BuyerRepository) GetPermissionsJobByID(jobID uuid.UUID) (*PermissionsJob, error) {
	var job PermissionsJob
	if err := r.db.Where("id = ?", jobID).First(&job).Error; err != nil {
		return nil, err
	}
	return &job, nil
}
