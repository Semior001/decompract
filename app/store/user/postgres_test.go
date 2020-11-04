package user

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Semior001/gotemplate/app/store"

	"github.com/jackc/pgx"
	"github.com/stretchr/testify/require"
)

func TestPostgres_GetPasswordHash(t *testing.T) {
	srv := preparePgStore(t)
	insertTestUsers(t, srv)

	pwd, err := srv.GetPasswordHash(usrs[1].ID)
	require.NoError(t, err)

	assert.Equal(t, usrs[1].Password, pwd)
}

func TestPostgres_GetUser(t *testing.T) {
	srv := preparePgStore(t)
	insertTestUsers(t, srv)

	// querying
	usr, err := srv.GetUser(usrs[3].ID)
	require.NoError(t, err)

	// checking
	shouldBe := usrs[3]
	assert.Equal(t, shouldBe.Email, usr.Email)
	assert.Equal(t, shouldBe.ID, usr.ID)

	assert.ElementsMatch(t, shouldBe.Privileges, usr.Privileges)
}

func TestPostgres_AddUser(t *testing.T) {
	srv := preparePgStore(t)

	id, err := srv.AddUser(store.User{
		ID:         "00000000-0000-0000-0000-000000000002",
		Email:      "foo@bar.com",
		Privileges: []store.Privilege{store.PrivAddUsers, store.PrivListUsers, store.PrivReadUsers},
	}, "blahblah", false)
	require.NoError(t, err)
	assert.Equal(t, "00000000-0000-0000-0000-000000000002", id)

	row := srv.connPool.QueryRow(`SELECT id, email, privileges, password FROM users`)
	var email, pwd string
	var privs []store.Privilege
	err = row.Scan(&id, &email, &privs, &pwd)
	require.NoError(t, err)

	id, err = srv.AddUser(store.User{
		ID:         "00000000-0000-0000-0000-000000000003",
		Email:      "foo@bar.com",
		Privileges: []store.Privilege{store.PrivListUsers, store.PrivReadUsers},
	}, "blahblah", true)
	require.NoError(t, err)
	// user with the same email already exists, it should ignore the new ID value and return
	// the ID of the older row
	assert.Equal(t, "00000000-0000-0000-0000-000000000002", id)

	row = srv.connPool.QueryRow(`SELECT id, email, privileges, password FROM users`)
	privs = []store.Privilege{}
	err = row.Scan(&id, &email, &privs, &pwd)
	require.NoError(t, err)
}

func insertTestUsers(t *testing.T, srv *Postgres) {
	tx, err := srv.connPool.Begin()
	require.NoError(t, err)
	defer func() {
		err := tx.Commit()
		require.NoError(t, err)
	}()

	for _, u := range usrs {
		_, err = srv.connPool.Exec("INSERT INTO users(id, email, password, privileges) "+
			"VALUES ($1, $2, $3, $4)", u.ID, u.Email, u.Password, u.Privileges)
		require.NoError(t, err)
	}
}

func preparePgStore(t *testing.T) *Postgres {
	// initializing connection with postgres
	connStr := os.Getenv("DB_TEST")
	connConf, err := pgx.ParseConnectionString(connStr)
	require.NoError(t, err)

	pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:     connConf,
		MaxConnections: 5,
		AcquireTimeout: time.Minute,
	})
	require.NoError(t, err)

	log.Printf("[INFO] initialized postgres connection pool to %s:%d", connConf.Host, connConf.Port)

	p, err := NewPostgres(pool, connConf)
	require.NoError(t, err)

	// setting cleanups
	cleanupStorage(t, p.connPool)
	t.Cleanup(func() {
		cleanupStorage(t, p.connPool)
	})

	return p
}

func cleanupStorage(t *testing.T, p *pgx.ConnPool) {
	tx, err := p.Begin()
	require.NoError(t, err)
	defer func() {
		err := tx.Commit()
		require.NoError(t, err)
	}()

	_, err = tx.Exec(`TRUNCATE users CASCADE`)
	require.NoError(t, err)
}

var usrs = []struct {
	ID         string
	Email      string
	Password   string
	Privileges []store.Privilege
}{
	{
		ID:         "00000000-0000-0000-0000-000000000001",
		Email:      "foo@bar.com",
		Password:   "blahblahblah",
		Privileges: nil,
	},
	{
		ID:         "00000000-0000-0000-0000-000000000002",
		Email:      "foo1@bar.com",
		Password:   "blahblahblah1",
		Privileges: []store.Privilege{store.PrivAddUsers, store.PrivEditUsers, store.PrivListUsers, store.PrivReadUsers},
	},
	{
		ID:         "00000000-0000-0000-0000-000000000003",
		Email:      "foo2@bar.com",
		Password:   "blahblahblah2",
		Privileges: []store.Privilege{store.PrivEditUsers},
	},
	{
		ID:         "00000000-0000-0000-0000-000000000004",
		Email:      "foo3@bar.com",
		Password:   "blahblahblah",
		Privileges: []store.Privilege{store.PrivAddUsers, store.PrivEditUsers, store.PrivListUsers, store.PrivReadUsers},
	},
	{
		ID:         "00000000-0000-0000-0000-000000000005",
		Email:      "foo4@bar.com",
		Password:   "blahblahblah",
		Privileges: []store.Privilege{store.PrivAddUsers},
	},
	{
		ID:         "00000000-0000-0000-0000-000000000006",
		Email:      "foo5@bar.com",
		Password:   "blahblahblah",
		Privileges: []store.Privilege{store.PrivReadUsers},
	},
}
