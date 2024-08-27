package domain

import "errors"

var (
	ErrEventConflict     = errors.New("event already exists")
	ErrOrderAlreadyFinal = errors.New("order already in final state")
	ErrOrderNotFound     = errors.New("order not found")
)

func IsDomainError(err error) bool {
	return errors.Is(err, ErrEventConflict) || errors.Is(err, ErrOrderAlreadyFinal) || errors.Is(err, ErrOrderNotFound)
}
