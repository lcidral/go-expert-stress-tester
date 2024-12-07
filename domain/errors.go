package domain

import "errors"

var (
	ErrTimeout    = errors.New("request timeout")
	ErrConnection = errors.New("connection error")
)
