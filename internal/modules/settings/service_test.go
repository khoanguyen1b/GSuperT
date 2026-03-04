package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSupportedSettingKeys_HasGPTKey(t *testing.T) {
	keys := SupportedSettingKeys()
	assert.Contains(t, keys, string(SettingKeyGPTAPIKey))
	assert.True(t, IsSupportedSettingKey(string(SettingKeyGPTAPIKey)))
	assert.False(t, IsSupportedSettingKey("unknown_key"))
}

func TestService_UpsertMany_EmptyPayload(t *testing.T) {
	svc := NewService(nil)
	_, err := svc.UpsertMany([]UpsertInput{})

	var validationErr *ValidationError
	assert.ErrorAs(t, err, &validationErr)
	assert.Equal(t, "at least one key-value pair is required", validationErr.Fields["body"])
}

func TestService_GetByKey_InvalidKey(t *testing.T) {
	svc := NewService(nil)
	_, err := svc.GetByKey("invalid_key")

	var validationErr *ValidationError
	assert.ErrorAs(t, err, &validationErr)
	assert.Contains(t, validationErr.Fields["key"], "unsupported key")
}
