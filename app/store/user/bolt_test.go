package user

import (
	"encoding/json"
	"github.com/Semior001/gotemplate/app/store"
	bolt "github.com/coreos/bbolt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestBolt_AddUser(t *testing.T) {
	svc := prepareBoltDB(t)
	u := store.User{
		ID:         "00000000-0000-0000-0000-000000000002",
		Email:      "foo@bar.com",
		Privileges: []store.Privilege{store.PrivAddUsers, store.PrivListUsers, store.PrivReadUsers},
	}
	pwd := "blahblah"

	checkUser := func(t *testing.T) {
		err := svc.db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(usersBktName)).Get([]byte(u.ID))
			assert.NotNil(t, b)

			var uu store.User
			err := json.Unmarshal(b, &uu)
			require.NoError(t, err)

			assert.Equal(t, u.ID, uu.ID)
			assert.Equal(t, u.Email, uu.Email)
			assert.ElementsMatch(t, u.Privileges, uu.Privileges)

			b = tx.Bucket([]byte(passwordsBktName)).Get([]byte(u.ID))
			assert.Equal(t, pwd, string(b))

			return nil
		})
		require.NoError(t, err)
	}

	id, err := svc.AddUser(u, pwd, false)
	require.NoError(t, err)
	assert.Equal(t, u.ID, id)
	checkUser(t)

	id, err = svc.AddUser(u, "blahblah1", true)
	require.NoError(t, err)
	assert.Equal(t, u.ID, id)
	checkUser(t)

	id, err = svc.AddUser(u, "blahblah2", false)
	require.Errorf(t, err, "user foo@bar.com already exists")
	checkUser(t)

}

func TestBolt_GetPasswordHash(t *testing.T) {
	svc := prepareBoltDB(t)
	err := svc.db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket([]byte(passwordsBktName)).Put(
			[]byte("00000000-0000-0000-0000-000000000002"),
			[]byte("verystrongpassword"),
		)
		return err
	})
	require.NoError(t, err)

	p, err := svc.GetPasswordHash("00000000-0000-0000-0000-000000000002")
	require.NoError(t, err)
	assert.Equal(t, "verystrongpassword", p)
}

func TestBolt_GetUser(t *testing.T) {
	svc := prepareBoltDB(t)
	u := store.User{
		ID:         "00000000-0000-0000-0000-000000000002",
		Email:      "foo@bar.com",
		Privileges: []store.Privilege{store.PrivAddUsers, store.PrivListUsers, store.PrivReadUsers},
	}

	err := svc.db.Update(func(tx *bolt.Tx) error {
		b, err := json.Marshal(u)
		require.NoError(t, err)

		err = tx.Bucket([]byte(usersBktName)).Put([]byte(u.ID), b)
		require.NoError(t, err)
		return nil
	})
	require.NoError(t, err)

	uu, err := svc.GetUser("00000000-0000-0000-0000-000000000002")
	require.NoError(t, err)
	assert.Equal(t, u.ID, uu.ID)
	assert.Equal(t, u.Email, uu.Email)
	assert.ElementsMatch(t, u.Privileges, uu.Privileges)
}

func prepareBoltDB(t *testing.T) *Bolt {
	loc, err := ioutil.TempDir("", "test_templ_users")
	require.NoError(t, err, "failed to make temp dir")

	svc, err := NewBoltStorage(path.Join(loc, "users_templ_test.db"), bolt.Options{})
	require.NoError(t, err, "New bolt storage")

	t.Cleanup(func() {
		assert.NoError(t, os.RemoveAll(loc))
	})
	return svc
}
