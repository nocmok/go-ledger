package model

type ErrorType string

const (
	ErrorTypeInternalError  ErrorType = "INTERNAL_ERROR"
	ErrorTypeInvalidRequest ErrorType = "INVALID_REQUEST"
)

type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Error struct {
	Type    ErrorType     `json:"type"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details"`
}

func (e *Error) Error() string {
	return e.Message
}
