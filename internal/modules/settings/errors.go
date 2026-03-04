package settings

import (
	"errors"
	"fmt"
	"strings"
)

var ErrSettingNotFound = errors.New("setting not found")

type ValidationError struct {
	Message string
	Fields  map[string]string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Fields)
}

func newUnsupportedKeyError(fieldPath string, key string) *ValidationError {
	return &ValidationError{
		Message: "validation failed",
		Fields: map[string]string{
			fieldPath: fmt.Sprintf(
				"unsupported key '%s'; supported keys: %s",
				key,
				strings.Join(SupportedSettingKeys(), ", "),
			),
		},
	}
}
