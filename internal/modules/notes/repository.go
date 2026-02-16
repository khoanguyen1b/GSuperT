package notes

import (
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(note *Note) error {
	return r.db.Create(note).Error
}

func (r *Repository) FindByID(id string) (*Note, error) {
	var note Note
	if err := r.db.First(&note, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &note, nil
}

func (r *Repository) Update(note *Note) error {
	return r.db.Save(note).Error
}

func (r *Repository) Delete(id string) error {
	return r.db.Delete(&Note{}, "id = ?", id).Error
}

func (r *Repository) List() ([]Note, error) {
	var notes []Note
	if err := r.db.Find(&notes).Error; err != nil {
		return nil, err
	}
	return notes, nil
}

func (r *Repository) ListByCustomerID(customerID string) ([]Note, error) {
	var notes []Note
	if err := r.db.Where("customer_id = ?", customerID).Find(&notes).Error; err != nil {
		return nil, err
	}
	return notes, nil
}
