package settings

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) UpsertMany(items []AppSetting) error {
	if len(items) == 0 {
		return nil
	}

	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "key"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"value":      gorm.Expr("EXCLUDED.value"),
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(&items).Error
}

func (r *Repository) FindByKey(key string) (*AppSetting, error) {
	var setting AppSetting
	if err := r.db.First(&setting, "key = ?", key).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *Repository) List() ([]AppSetting, error) {
	var settings []AppSetting
	if err := r.db.Order("key ASC").Find(&settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}
