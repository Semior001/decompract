package service

import (
	"golang.org/x/crypto/bcrypt"
	"testing"

	"github.com/Semior001/gotemplate/app/store/user"

	"github.com/Semior001/gotemplate/app/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataStore_AddAndRegisterUser(t *testing.T) {
	privs := []store.Privilege{store.PrivAddUsers, store.PrivEditUsers, store.PrivListUsers, store.PrivReadUsers}
	expected := store.User{
		Email:      "foo@bar.com",
		Privileges: []store.Privilege{store.PrivAddUsers, store.PrivEditUsers, store.PrivListUsers, store.PrivReadUsers},
	}
	expectedPwd := "some very strong password"

	// add user
	ur := &user.InterfaceMock{AddUserFunc: func(user store.User, pwd string, ignoreIfExists bool) (string, error) {
		assert.NotEmpty(t, user.ID)

		// to make structs be equal to user the single assert.Equal
		expected.ID = user.ID

		// as the order of privileges may be different
		assert.ElementsMatch(t, privs, user.Privileges)
		expected.Privileges = nil
		user.Privileges = nil

		assert.Equal(t, expected, user)

		err := bcrypt.CompareHashAndPassword([]byte(pwd), []byte(expectedPwd))
		require.NoError(t, err)
		assert.False(t, ignoreIfExists)
		return expected.ID, nil
	}}
	srv := DataStore{UserRepository: ur, BCryptCost: 4}

	id, err := srv.AddUser(expected, expectedPwd)
	require.NoError(t, err)
	assert.Equal(t, expected.ID, id)

	// register admin
	expected.Privileges = privs
	ur.AddUserFunc = func(user store.User, pwd string, ignoreIfExists bool) (string, error) {
		assert.NotEmpty(t, user.ID)

		// to make structs be equal to user the single assert.Equal
		expected.ID = user.ID

		// as the order of privileges may be different
		assert.ElementsMatch(t, privs, user.Privileges)
		user.Privileges = nil
		expected.Privileges = nil

		assert.Equal(t, expected, user)

		err := bcrypt.CompareHashAndPassword([]byte(pwd), []byte(expectedPwd))
		require.NoError(t, err)
		assert.True(t, ignoreIfExists)
		return expected.ID, nil
	}

	id, err = srv.RegisterAdmin(expected.Email, expectedPwd)
	require.NoError(t, err)
	assert.Equal(t, expected.ID, id)
}

func TestDataStore_PassThroughMethods(t *testing.T) {
	usr := store.User{
		ID:         "some awesome userID",
		Email:      "foo@bar.com",
		Privileges: []store.Privilege{store.PrivReadUsers, store.PrivListUsers},
	}

	srv := DataStore{UserRepository: &user.InterfaceMock{
		GetUserFunc: func(id string) (store.User, error) {
			assert.Equal(t, usr.ID, id)
			return usr, nil
		},
	}, BCryptCost: 4}

	email, err := srv.GetUserEmail(usr.ID)
	require.NoError(t, err)
	assert.Equal(t, usr.Email, email)

	privs, err := srv.GetUserPrivs(usr.ID)
	require.NoError(t, err)
	assert.Equal(t, usr.Privileges, privs)
}

func TestDataStore_CheckUserCredentials(t *testing.T) {
	b, err := bcrypt.GenerateFromPassword([]byte("some very protected pwd"), 4)
	require.NoError(t, err)

	srv := DataStore{UserRepository: &user.InterfaceMock{
		GetPasswordHashFunc: func(email string) (string, error) {
			assert.Equal(t, generateUserID("foo@bar.cc"), email)
			return string(b), err
		},
	}}
	ok, err := srv.CheckUserCredentials("foo@bar.cc", "some very protected pwd")
	require.NoError(t, err)
	assert.True(t, ok)
}
