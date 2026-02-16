package users

import (
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(user *User) error {
	return r.db.Create(user).Error
}

func (r *Repository) FindByID(id string) (*User, error) {
	var user User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) FindByEmail(email string) (*User, error) {
	var user User
	if err := r.db.First(&user, "email = ?", email).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) Update(user *User) error {
	return r.db.Save(user).Error
}

func (r *Repository) Delete(id string) error {
	return r.db.Delete(&User{}, "id = ?", id).Error
}

func (r *Repository) List() ([]User, error) {
	var users []User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// Refresh Token Repository Methods
func (r *Repository) CreateRefreshToken(rt *RefreshToken) error {
	return r.db.Create(rt).Error
}

func (r *Repository) FindRefreshToken(tokenHash string) (*RefreshToken, error) {
	var rt RefreshToken
	if err := r.db.First(&rt, "token_hash = ?", tokenHash).Error; err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *Repository) DeleteRefreshToken(userID string) error {
	return r.db.Delete(&RefreshToken{}, "user_id = ?", userID).Error
}
