package customers

import (
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(customer *Customer) error {
	return r.db.Create(customer).Error
}

func (r *Repository) FindByID(id string) (*Customer, error) {
	var customer Customer
	if err := r.db.First(&customer, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *Repository) Update(customer *Customer) error {
	return r.db.Save(customer).Error
}

func (r *Repository) Delete(id string) error {
	return r.db.Delete(&Customer{}, "id = ?", id).Error
}

func (r *Repository) List() ([]Customer, error) {
	var customers []Customer
	if err := r.db.Find(&customers).Error; err != nil {
		return nil, err
	}
	return customers, nil
}
