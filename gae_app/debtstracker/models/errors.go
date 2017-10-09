package models

import "github.com/pkg/errors"

var (
	ErrJsonCountMismatch = errors.New("json slice length is different to length of corresponding count property")
)
