package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Semior001/decompract/app/solver"

	"github.com/go-chi/render"

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
	WebRoot string

	Solvers []solver.Interface

	httpServer *http.Server
	lock       sync.Mutex
}

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
	rest.SendErrorHTML(w, r, http.StatusNotFound, nil, "not found")
}

func (s *Rest) routes() chi.Router {
	r := chi.NewRouter()

	r.Use(R.AppInfo("decompract", "Semior001", s.Version))
	r.Use(R.Recoverer(log.Default()))
	r.Use(R.Ping, middleware.RealIP)
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	r.NotFound(s.notFound)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Timeout(5 * time.Second))
	})

	r.Route("/api", func(rapi chi.Router) {
		rapi.Post("/plot", s.plotGraphsCtrl)
	})

	addFileServer(r, "/", http.Dir(s.WebRoot))

	return r
}

// GET /api/plot - plot graphs according to the given parameters
func (s *Rest) plotGraphsCtrl(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		rest.SendErrorHTML(w, r, http.StatusBadRequest, err, "failed to parse form data")
		return
	}

	var x0, y0, xEnd, n float64
	if len(r.Form["x0"]) != 1 || len(r.Form["y0"]) != 1 || len(r.Form["x_end"]) != 1 || len(r.Form["n"]) != 1 {
		rest.SendErrorHTML(w, r, http.StatusBadRequest, errors.New("invalid request"),
			"some fields are empty or contains more or less entries, than needed")
		return
	}
	if err := json.Unmarshal([]byte(r.Form["x0"][0]), &x0); err != nil {
		rest.SendErrorHTML(w, r, http.StatusForbidden, err, "can't read x0")
		return
	}
	if err := json.Unmarshal([]byte(r.Form["y0"][0]), &y0); err != nil {
		rest.SendErrorHTML(w, r, http.StatusForbidden, err, "can't read y0")
		return
	}
	if err := json.Unmarshal([]byte(r.Form["x_end"][0]), &xEnd); err != nil {
		rest.SendErrorHTML(w, r, http.StatusForbidden, err, "can't read xEnd")
		return
	}
	if err := json.Unmarshal([]byte(r.Form["n"][0]), &n); err != nil {
		rest.SendErrorHTML(w, r, http.StatusForbidden, err, "can't read n")
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"x0":   x0,
		"y0":   y0,
		"xend": xEnd,
		"n":    n,
	})
}

func addFileServer(r chi.Router, path string, root http.FileSystem) {
	log.Printf("[INFO] run file server for %s, path %s", root, path)
	webFS := http.FileServer(root)

	origPath := path
	webFS = http.StripPrefix(path, webFS)
	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		// don't show dirs, just serve files
		if strings.HasSuffix(r.URL.Path, "/") && len(r.URL.Path) > 1 && r.URL.Path != (origPath+"/") {
			http.NotFound(w, r)
			return
		}
		webFS.ServeHTTP(w, r)
	})
}
