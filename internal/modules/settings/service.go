package settings

import (
	"errors"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) UpsertMany(inputs []UpsertInput) ([]AppSetting, error) {
	if len(inputs) == 0 {
		return nil, &ValidationError{
			Message: "validation failed",
			Fields: map[string]string{
				"body": "at least one key-value pair is required",
			},
		}
	}

	lastByKey := make(map[string]string, len(inputs))
	keys := make([]string, 0, len(inputs))
	seen := make(map[string]struct{}, len(inputs))

	for i, input := range inputs {
		key := strings.ToLower(strings.TrimSpace(input.Key))
		if key == "" {
			return nil, &ValidationError{
				Message: "validation failed",
				Fields: map[string]string{
					"items[" + strconv.Itoa(i) + "].key": "key is required",
				},
			}
		}
		if !IsSupportedSettingKey(key) {
			return nil, newUnsupportedKeyError("items["+strconv.Itoa(i)+"].key", key)
		}

		lastByKey[key] = input.Value
		if _, ok := seen[key]; !ok {
			keys = append(keys, key)
			seen[key] = struct{}{}
		}
	}

	items := make([]AppSetting, 0, len(keys))
	for _, key := range keys {
		items = append(items, AppSetting{
			Key:   key,
			Value: lastByKey[key],
		})
	}

	if err := s.repo.UpsertMany(items); err != nil {
		return nil, err
	}

	result := make([]AppSetting, 0, len(items))
	for _, key := range keys {
		setting, err := s.repo.FindByKey(key)
		if err != nil {
			return nil, err
		}
		result = append(result, *setting)
	}

	return result, nil
}

func (s *Service) List() ([]AppSetting, error) {
	return s.repo.List()
}

func (s *Service) GetByKey(key string) (*AppSetting, error) {
	normalizedKey := strings.ToLower(strings.TrimSpace(key))
	if !IsSupportedSettingKey(normalizedKey) {
		return nil, newUnsupportedKeyError("key", normalizedKey)
	}

	setting, err := s.repo.FindByKey(normalizedKey)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSettingNotFound
		}
		return nil, err
	}

	return setting, nil
}
