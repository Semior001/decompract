package rest

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Semior001/gotemplate/app/store"
	"github.com/stretchr/testify/assert"
)

func TestUser_GetUserOrEmpty(t *testing.T) {
	r, err := http.NewRequest("GET", "http://blah.com", nil)
	require.NoError(t, err)
	u := GetUserOrEmpty(r)
	assert.Equal(t, store.User{}, u)

	r = SetUserInfo(r, store.User{ID: "blah", Email: "blah@blah.com", Privileges: []store.Privilege{store.PrivReadUsers}})
	u = GetUserOrEmpty(r)
	require.NoError(t, err)
	assert.Equal(t, store.User{ID: "blah", Email: "blah@blah.com", Privileges: []store.Privilege{store.PrivReadUsers}}, u)
}

func TestUser_GetUserInfo(t *testing.T) {
	r, err := http.NewRequest("GET", "http://blah.com", nil)
	require.NoError(t, err)
	_, err = GetUserInfo(r)
	assert.Error(t, err, "no user info")

	r = SetUserInfo(r, store.User{ID: "blah", Email: "blah@blah.com", Privileges: []store.Privilege{store.PrivReadUsers}})
	u, err := GetUserInfo(r)
	require.NoError(t, err)
	assert.Equal(t, store.User{ID: "blah", Email: "blah@blah.com", Privileges: []store.Privilege{store.PrivReadUsers}}, u)
}

func TestUser_MustGetUserInfo(t *testing.T) {
	r, err := http.NewRequest("GET", "http://blah.com", nil)
	require.NoError(t, err)
	assert.Panics(t, func() { _ = MustGetUserInfo(r) }, "should panic")

	r = SetUserInfo(r, store.User{ID: "blah", Email: "blah@blah.com", Privileges: []store.Privilege{store.PrivReadUsers}})
	u := MustGetUserInfo(r)
	require.NoError(t, err)
	assert.Equal(t, store.User{ID: "blah", Email: "blah@blah.com", Privileges: []store.Privilege{store.PrivReadUsers}}, u)
}
