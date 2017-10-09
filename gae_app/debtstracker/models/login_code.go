package models

import (
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

const LoginCodeKind = "LoginCode"

type LoginCodeEntity struct {
	Created time.Time
	Claimed time.Time
	UserID  int64
}

const CODE_LENGTH = 5

func LoginCodeToString(code int32) string {
	return fmt.Sprintf("%0"+strconv.Itoa(CODE_LENGTH)+"d", code)
}

var (
	ErrLoginCodeExpired        = errors.New("Code expired")         // TODO: Show we move this to DAL?
	ErrLoginCodeAlreadyClaimed = errors.New("Code already claimed") // TODO: Show we move this to DAL?
)
