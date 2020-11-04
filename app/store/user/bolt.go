package user

import (
	"encoding/json"
	"fmt"
	"github.com/Semior001/gotemplate/app/store"
	bolt "github.com/coreos/bbolt"
	log "github.com/go-pkgz/lgr"
	"github.com/pkg/errors"
)

const usersBktName = "users"
const passwordsBktName = "user_passwords"

// Bolt implements Interface to contain, fetch and update users
// there is one top-level bucket:
// - users with the k:v pair as userID:user
// - user passwords with the k:v pair as userID:password
type Bolt struct {
	fileName string
	db       *bolt.DB
}

// NewBoltStorage creates buckets and initial data processing
func NewBoltStorage(fileName string, options bolt.Options) (*Bolt, error) {
	db, err := bolt.Open(fileName, 0600, &options)
	if err != nil {
		return nil, fmt.Errorf("failed to make boltdb for %s: %w", fileName, err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(usersBktName)); err != nil {
			return fmt.Errorf("failed to create top-level bucket %s: %w", usersBktName, err)
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(passwordsBktName)); err != nil {
			return fmt.Errorf("failed to create top-level bucket %s: %w", passwordsBktName, err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize boltdb buckets for %s: %w", fileName, err)
	}

	log.Printf("[INFO] Users BoltDB instantiated")
	return &Bolt{
		db:       db,
		fileName: fileName,
	}, nil
}

// GetUser by its id
func (b *Bolt) GetUser(id string) (store.User, error) {
	var u store.User
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(usersBktName)).Get([]byte(id))
		if b == nil {
			return ErrNotFound
		}
		return errors.Wrapf(json.Unmarshal(b, &u), "failed to unmarshal user %s", id)
	})
	return u, err
}

// GetPasswordHash of the user by its email
func (b *Bolt) GetPasswordHash(id string) (pwd string, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(passwordsBktName)).Get([]byte(id))
		if b == nil {
			return ErrNotFound
		}
		pwd = string(b)
		return nil
	})
	return pwd, err
}

// AddUser to the storage
func (b *Bolt) AddUser(user store.User, pwd string, ignoreIfExists bool) (string, error) {
	if user.ID == "" {
		return "", errors.Errorf("id for user %s is not assigned", user.Email)
	}
	usrExists := func(tx *bolt.Tx) bool {
		bu := tx.Bucket([]byte(usersBktName)).Get([]byte(user.ID))
		bp := tx.Bucket([]byte(passwordsBktName)).Get([]byte(user.ID))
		return bu != nil || bp != nil
	}
	err := b.db.Update(func(tx *bolt.Tx) error {
		// checking that user is not yet in the bucket
		if usrExists(tx) {
			if ignoreIfExists {
				return nil
			}
			return errors.Errorf("user %s already exists", user.Email)
		}

		// adding user to users bucket
		b, err := json.Marshal(user)
		if err != nil {
			return errors.Wrap(err, "failed to marshal")
		}

		if err = tx.Bucket([]byte(usersBktName)).Put([]byte(user.ID), b); err != nil {
			return errors.Wrapf(err, "failed to put user %s into its bucket", user.Email)
		}

		// adding user password
		if err = tx.Bucket([]byte(passwordsBktName)).Put([]byte(user.ID), []byte(pwd)); err != nil {
			return errors.Wrapf(err, "failed to put user's password into its bucket, user: %s", user.Email)
		}

		return nil
	})
	return user.ID, err
}
