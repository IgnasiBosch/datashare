package core

type Error struct {
	Status int    `json:"status"`
	Code   int    `json:"code"`
	ErrMsg string `json:"msg"`
}

// Error method implementation for core.Error.
// This makes core.Error adhere to the builtin.error interface.
func (e *Error) Error() string {
	return e.ErrMsg
}

func NewError(status int, code int, message string) *Error {
	return &Error{
		Status: status,
		Code:   code,
		ErrMsg: message,
	}
}

type IDKey struct {
	ID  string `json:"id"`
	Key string `json:"key"`
}

func NewIDKey(ID string, key string) *IDKey {
	return &IDKey{
		ID:  ID,
		Key: key,
	}
}
