package database

import (
	"errors"
	"fmt"
)

type ErrDublicateKey struct {
	Key string
}

func (m ErrDublicateKey) Error() string {
	return fmt.Sprintf("ERROR: duplicate key value violates unique constraint %q (SQLSTATE 23505)", m.Key)
}

var (
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrCreatedBySameUser   = errors.New("order was already created by the same user")
	ErrCreatedDiffUser     = errors.New("order was already created by the other user")
	ErrNoData              = errors.New("no data")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrNoSuchUser          = errors.New("no such user")
	ErrInvalidCredentials  = errors.New("incorrect password")
)
