package topic_conversation

import "fmt"

const (
	MinTurns     = 20
	MaxTurns     = 100
	TurnStep     = 10
	DefaultTurns = 20
)

type ValidationError struct {
	Message string
	Fields  map[string]string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Fields)
}

type ProviderError struct {
	Message string
}

func (e *ProviderError) Error() string {
	return e.Message
}

func ValidateGenerateRequest(req *GenerateRequest) error {
	fields := map[string]string{}
	if req == nil {
		fields["request"] = "request body is required"
		return &ValidationError{
			Message: "validation failed",
			Fields:  fields,
		}
	}

	if req.Turns != 0 {
		if req.Turns < MinTurns || req.Turns > MaxTurns {
			fields["turns"] = fmt.Sprintf("turns must be between %d and %d", MinTurns, MaxTurns)
		} else if req.Turns%TurnStep != 0 {
			fields["turns"] = fmt.Sprintf("turns must be a multiple of %d (e.g. 20, 30, ..., 100)", TurnStep)
		}
	}

	if len(fields) > 0 {
		return &ValidationError{
			Message: "validation failed",
			Fields:  fields,
		}
	}

	return nil
}
