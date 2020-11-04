package api

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Semior001/decompract/app/rest"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/httprate"
	log "github.com/go-pkgz/lgr"
	R "github.com/go-pkgz/rest"
)

// Rest defines a simple web server for routing to calendar REST api methods
type Rest struct {
	Version string

	httpServer *http.Server
	lock       sync.Mutex

	// todo ctrl groups
}

const hardBodyLimit = 1024 * 64 //nolint // limit size of body

// Run starts the web-server for listening
func (s *Rest) Run(port int) {
	s.lock.Lock()
	s.httpServer = s.makeHTTPServer(port, s.routes())
	s.httpServer.ErrorLog = log.ToStdLogger(log.Default(), "WARN")
	s.lock.Unlock()

	log.Printf("[INFO] started web server at port %d", port)
	err := s.httpServer.ListenAndServe()
	log.Printf("[WARN] web server terminated reason: %s", err)
}

func (s *Rest) makeHTTPServer(port int, routes chi.Router) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           routes,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
}

// notFound returns standard 404 not found message
func (s *Rest) notFound(w http.ResponseWriter, r *http.Request) {
	rest.SendErrorJSON(w, r, http.StatusNotFound, nil, "not found", rest.ErrBadRequest)
}

func (s *Rest) controllerGroups() {
	// todo create ctrl groups
}

func (s *Rest) routes() chi.Router {
	r := chi.NewRouter()

	r.Use(R.AppInfo("decompract", "Semior001", s.Version))
	r.Use(R.Recoverer(log.Default()))
	r.Use(R.Ping, middleware.RealIP)
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	r.NotFound(s.notFound)

	// todo init ctrl groups and insert them into Rest
	s.controllerGroups()

	r.Group(func(r chi.Router) {
		r.Use(middleware.Timeout(5 * time.Second))
	})

	r.Route("/api/v1", func(rapi chi.Router) {
		// todo mount and give patterns for ctrl groups
	})

	return r
}
