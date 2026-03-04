package settings

import "sort"

type SettingKey string

const (
	SettingKeyGPTAPIKey SettingKey = "gpt_api_key"
)

var supportedKeys = map[string]struct{}{
	string(SettingKeyGPTAPIKey): {},
}

func IsSupportedSettingKey(key string) bool {
	_, ok := supportedKeys[key]
	return ok
}

func SupportedSettingKeys() []string {
	keys := make([]string, 0, len(supportedKeys))
	for key := range supportedKeys {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
