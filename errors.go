package chunk_upload

import "fmt"

type Error struct {
	Err    errKind `json:"error"`
	Detail string  `json:"detail"`
}

type errKind string

const (
	ErrNotFound         errKind = "not_found"
	ErrDisMatch         errKind = "dismatch"
	ErrparameterInvalid errKind = "parameterInvalid"
)

func (e Error) Error() string {
	return fmt.Sprintf("%s:%s", e.Err, e.Detail)
}

func newError(kind errKind, msg string) Error {
	return Error{
		Err:    kind,
		Detail: msg,
	}
}
