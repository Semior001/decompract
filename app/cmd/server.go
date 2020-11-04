package cmd

import (
	"strings"
	"time"

	"github.com/Semior001/gotemplate/app/store"

	"github.com/pkg/errors"

	"github.com/go-pkgz/auth/provider"

	"github.com/Semior001/gotemplate/app/store/service"

	"github.com/go-pkgz/auth/avatar"

	"github.com/go-pkgz/auth/token"

	"github.com/Semior001/gotemplate/app/rest/api"
	"github.com/go-pkgz/auth"
	log "github.com/go-pkgz/lgr"
)

// Server runs REST API web server
type Server struct {
	ServiceURL string `long:"service_url" env:"SERVICE_URL" description:"http service url" required:"true"`
	Port       int    `long:"service_port" env:"SERVICE_PORT" description:"http server port" default:"8080"`

	Auth struct {
		TTL struct {
			JWT    time.Duration `long:"jwt" env:"JWT" default:"5m" description:"jwt TTL"`
			Cookie time.Duration `long:"cookie" env:"COOKIE" default:"200h" description:"auth cookie TTL"`
		} `group:"ttl" namespace:"ttl" env-namespace:"TTL"`
		Secret     string `long:"secret" env:"SECRET" description:"secret for authentication tokens"`
		BCryptCost int    `long:"bcrypt_cost" env:"BCRYPT_COST" description:"bcrypt cost for hashing user password" default:"10"`
	} `group:"auth" namespace:"auth" env-namespace:"AUTH"`
	DBConnStr string `long:"db_conn_str" env:"DB_CONN_STR" required:"true" description:"connection string to db"`

	Admin AdminGroup `group:"admin" namespace:"admin" env-namespace:"ADMIN"`

	CommonOpts
}

// AdminGroup defines options group for admin params
type AdminGroup struct {
	Email    string `long:"email" env:"EMAIL" description:"default admin email" required:"true"`
	Password string `long:"password" env:"PASSWORD" description:"default admin password" required:"true"`
}

// Execute runs http web server
func (s *Server) Execute(_ []string) error {
	ds, err := prepareDataStore(s.DBConnStr, s.Auth.BCryptCost)
	if err != nil {
		return errors.Wrapf(err, "failed to prepare data store with %s and cost %d", s.DBConnStr, s.Auth.BCryptCost)
	}

	adminID, err := ds.RegisterAdmin(s.Admin.Email, s.Admin.Password)
	if err != nil {
		return errors.Wrapf(err, "failed to register admin %s:%s", s.Admin.Email, s.Admin.Password)
	}
	log.Printf("[DEBUG] registered admin %s with id %s", s.Admin.Email, adminID)

	authenticator := s.makeAuthenticator(ds)
	srv := api.Rest{
		Version:       s.Version,
		Authenticator: authenticator,
		DataStore:     ds,
	}

	srv.Run(s.Port)
	return nil
}

func (s *Server) makeAuthenticator(ds *service.DataStore) *auth.Service {
	authenticator := auth.NewService(auth.Opts{
		SecretReader: token.SecretFunc(func(aud string) (string, error) {
			return s.Auth.Secret, nil
		}),
		ClaimsUpd: token.ClaimsUpdFunc(func(c token.Claims) token.Claims { // set attributes, on new token or refresh
			if c.User == nil {
				return c
			}

			var err error
			c.User.Email, err = ds.GetUserEmail(c.User.ID)
			if err != nil {
				log.Printf("[WARN] can't read email for %s, %v", c.User.ID, err)
			}

			privs, err := ds.GetUserPrivs(c.User.ID)
			if err != nil {
				log.Printf("[WARN] can't get privs for %s, %v ", c.User.ID, err)
			}

			c.User.SetSliceAttr("privileges", store.PrivsToStr(privs))

			return c
		}),
		SecureCookies:  strings.HasPrefix(s.ServiceURL, "https://"),
		TokenDuration:  s.Auth.TTL.JWT,
		CookieDuration: s.Auth.TTL.Cookie,
		DisableXSRF:    true,
		Issuer:         "gotemplate",
		URL:            strings.TrimSuffix(s.ServiceURL, "/"),
		Validator: token.ValidatorFunc(func(token string, claims token.Claims) bool { // check on each auth call (in middleware)
			return claims.User != nil
		}),
		AvatarStore: avatar.NewNoOp(),
		Logger:      log.Default(),
	})
	authenticator.AddDirectProvider("local", provider.CredCheckerFunc(ds.CheckUserCredentials))
	return authenticator
}
