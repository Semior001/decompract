// Package user provides implementations for Interface for the database user repository.
// All consumers should be using service.DataStore and not the naked repositories!
package user

import (
	"errors"
	"github.com/Semior001/gotemplate/app/store"
)

//go:generate moq -out mock_user.go . Interface

// Store-defined errors
var (
	ErrNotFound = errors.New("user not found")
)

// Interface defines methods to repository, and fetch models
type Interface interface {
	GetUser(id string) (u store.User, err error)
	GetPasswordHash(email string) (pwd string, err error)
	AddUser(user store.User, pwd string, ignoreIfExists bool) (id string, err error)
}
