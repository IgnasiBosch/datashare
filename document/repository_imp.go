package document

import (
	"gorm.io/gorm"
	"time"
)

type RepositoryImp struct {
	Db *gorm.DB
}

func NewRepositoryImp(db *gorm.DB) *RepositoryImp {
	return &RepositoryImp{Db: db}
}

func (r *RepositoryImp) FindById(id string) (*Document, error) {
	var d Document
	err := r.Db.First(&d, "id = ?", id).Error
	return &d, err
}

func (r *RepositoryImp) Save(d *Document) error {
	return r.Db.Create(&d).Error
}

func (r *RepositoryImp) Update(d *Document) error {
	return r.Db.Save(&d).Error
}

func (r *RepositoryImp) GetExpired() ([]Document, error) {
	var documents []Document
	uploadedBefore := time.Now().Add(-time.Hour * 24)
	err := r.Db.Find(&documents, "status = ? AND uploaded_at < ?", Ready, uploadedBefore).Error
	return documents, err
}

func (r *RepositoryImp) GetTotalUsage(client string) (int64, error) {
	var total int64
	uploadedBefore := time.Now().Add(-time.Hour)
	err := r.Db.Model(&Document{}).Where("client = ? AND uploaded_at >= ?", client, uploadedBefore).Count(&total).Error
	return total, err
}
