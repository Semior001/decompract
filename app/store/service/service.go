// Package service wraps user interfaces with common logic unrelated to any particular user implementation.
// All consumers should be using service.DataStore and not the naked repositories!
package service

import (
	"crypto/sha1" // nolint
	"log"

	"github.com/go-pkgz/auth/token"

	"github.com/Semior001/gotemplate/app/store"
	"golang.org/x/crypto/bcrypt"

	"github.com/Semior001/gotemplate/app/store/user"

	"github.com/pkg/errors"
)

// DataStore wraps all stores with common and additional methods
// todo looks ugly, rewrite
type DataStore struct {
	UserRepository user.Interface
	BCryptCost     int
}

// GetUserEmail returns the email of the specified user
func (s *DataStore) GetUserEmail(id string) (email string, err error) {
	u, err := s.UserRepository.GetUser(id)
	//goland:noinspection GoNilness
	return u.Email, errors.Wrapf(err, "failed to read email of %s", id)
}

// GetUserPrivs returns the list of privileges of the specified user
func (s *DataStore) GetUserPrivs(id string) (privs []store.Privilege, err error) {
	u, err := s.UserRepository.GetUser(id)
	//goland:noinspection GoNilness
	return u.Privileges, errors.Wrapf(err, "failed to read privs of %s", id)
}

// CheckUserCredentials with the given username and password
func (s *DataStore) CheckUserCredentials(email string, password string) (ok bool, err error) {
	userpwd, err := s.UserRepository.GetPasswordHash(generateUserID(email))
	if err != nil {
		return false, errors.Wrapf(err, "failed to validate user")
	}
	err = bcrypt.CompareHashAndPassword([]byte(userpwd), []byte(password))
	return err == nil, err
}

// AddUser to the database, hash its password and give it an ID, if needed
func (s *DataStore) AddUser(user store.User, password string) (id string, err error) {
	// hashing password
	b, err := bcrypt.GenerateFromPassword([]byte(password), s.BCryptCost)
	if err != nil {
		return "", errors.Wrapf(err, "failed to hash %s user's password with bcrypt", user.Email)
	}
	// adding id
	if user.ID == "" {
		user.ID = generateUserID(user.Email)
	}

	id, err = s.UserRepository.AddUser(user, string(b), false)
	return id, errors.Wrapf(err, "failed to add user %s to database", user.ID)
}

// RegisterAdmin in the database
func (s *DataStore) RegisterAdmin(email string, password string) (id string, err error) {
	// hashing password
	b, err := bcrypt.GenerateFromPassword([]byte(password), s.BCryptCost)
	if err != nil {
		return "", errors.Wrapf(err, "failed to hash %s user's password with bcrypt", email)
	}
	u := store.User{
		ID:         generateUserID(email),
		Email:      email,
		Privileges: []store.Privilege{store.PrivReadUsers, store.PrivEditUsers, store.PrivListUsers, store.PrivAddUsers},
	}
	log.Printf("[INFO] trying to register admin with %+v and pwd %s", u, password)
	id, err = s.UserRepository.AddUser(u, string(b), true)
	return id, errors.Wrapf(err, "failed to add user %s to database", u.ID)
}

func generateUserID(email string) string {
	return "local_" + token.HashID(sha1.New(), email) //nolint
}
