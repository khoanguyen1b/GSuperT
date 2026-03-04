package settings

import "time"

type AppSetting struct {
	Key       string    `gorm:"type:varchar(100);primaryKey" json:"key"`
	Value     string    `gorm:"type:text;not null" json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (AppSetting) TableName() string {
	return "app_settings"
}

type UpsertInput struct {
	Key   string
	Value string
}
