package domain

import "errors"

var (
	ErrEventConflict     = errors.New("event already exists")
	ErrOrderAlreadyFinal = errors.New("order already final")
)
