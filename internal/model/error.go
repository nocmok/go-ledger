package model

type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Error struct {
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details"`
}

func (e *Error) Error() string {
	return e.Message
}
