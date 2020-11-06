package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"gonum.org/v1/plot/plotter"

	"github.com/Semior001/decompract/app/graph"

	"github.com/pkg/errors"

	"github.com/Semior001/decompract/app/solver"

	"github.com/go-chi/render"

	"github.com/Semior001/decompract/app/rest"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/httprate"
	log "github.com/go-pkgz/lgr"
	R "github.com/go-pkgz/rest"
)

const plotHTMLTmpl = `<!DOCTYPE html>
<html>
<head>
    <meta name="viewport" content="width=device-width"/>
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8"/>
    <title>Solution</title>
</head>
<body>
<div style="text-align: center; font-family: Arial, sans-serif; font-size: 18px;">
    <h1 style="position: relative; color: #4fbbd6; margin-top: 0.2em;">DEComPract</h1>
    <h3 style="position: relative; color: #666666; margin-top: 0.2em;">Yelshat Duskaliyev, B19-04</h3>
    <p>x<sub>0</sub>={{printf "%.4f" .X0}}; y<sub>0</sub>={{printf "%.4f" .Y0}}; X = {{printf "%.4f" .XEnd}}; N = {{.N}}</p>
    <a href="/">Enter another data</a>
</div>
<table width="100%" style="align-content: center; font-family: Arial, sans-serif; font-size: 18px; position: relative; margin-top: 0.2em;">
    <tr>
        <td><img width="100%" src="data:image/jpg;base64,{{.SolutionsImg}}" alt="solutions plot"></td>
        <td><img width="100%" src="data:image/jpg;base64,{{.LTEImg}}" alt="lte plot"></td>
        <td><img width="100%" src="data:image/jpg;base64,{{.GTEImg}}" alt="gte plot"></td>
    </tr>
</table>
</body>
</html>`

type plotTmplData struct {
	X0           float64
	Y0           float64
	XEnd         float64
	N            int
	SolutionsImg string
	LTEImg       string
	GTEImg       string
}

// Rest defines a simple web server for routing to calendar REST api methods
type Rest struct {
	Version string
	WebRoot string

	Solvers     []solver.Interface
	ExactSolver *solver.Exact
	Plotter     graph.Plotter

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

	addFileServer(r, "/", http.Dir(s.WebRoot))
	r.Post("/", s.plotGraphsCtrl)

	return r
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

// GET /api/plot - plot graphs according to the given parameters
func (s *Rest) plotGraphsCtrl(w http.ResponseWriter, r *http.Request) {
	// reading form
	if err := r.ParseForm(); err != nil {
		rest.SendErrorHTML(w, r, http.StatusBadRequest, err, "failed to parse form data")
		return
	}
	req, err := readVals(r.Form)
	if err != nil {
		rest.SendErrorHTML(w, r, http.StatusForbidden, err, "failed to read request values")
		return
	}

	var lines []graph.Line

	// solving equation by the given solvers
	for _, slvr := range s.Solvers {
		name := slvr.Name()

		var pts []plotter.XY
		err := slvr.Solve(solver.CalculateStepSize(req.N, req.X0, req.XEnd), req.X0, req.Y0, req.XEnd,
			solver.DrawerFunc(func(ps solver.Point) error {
				pts = append(pts, ps.XY())
				return nil
			}),
		)
		if err != nil {
			rest.SendErrorHTML(w, r, http.StatusBadRequest, err, fmt.Sprintf("can't solve with %s", name))
			return
		}

		lines = append(lines, graph.Line{Name: name, Points: pts})
	}

	// adding exact solution
	var exactPts []plotter.XY
	err = s.ExactSolver.Solve(solver.CalculateStepSize(req.N, req.X0, req.XEnd), req.X0, req.Y0, req.XEnd,
		solver.DrawerFunc(func(ps solver.Point) error {
			exactPts = append(exactPts, ps.XY())
			return nil
		}),
	)
	if err != nil {
		rest.SendErrorHTML(w, r, http.StatusBadRequest, err, "can't solve with exact")
		return
	}

	// calculating errors for solvers
	var lte []graph.Line
	for _, line := range lines {
		var pts plotter.XYs
		for i := range exactPts {
			y := math.Abs(line.Points[i].Y - exactPts[i].Y)
			pts = append(pts, plotter.XY{X: exactPts[i].X, Y: y})
		}
		lte = append(lte, graph.Line{Name: line.Name, Points: pts})
	}

	// plotting lte graph
	b, err := s.Plotter.Plot("LTE", lte)
	if err != nil {
		rest.SendErrorHTML(w, r, http.StatusInternalServerError, err, "can't plot the lte graph")
		return
	}
	b64LTEGraph := base64.StdEncoding.EncodeToString(b)

	// plotting gte max
	var gte []graph.Line
	for _, slvr := range s.Solvers {
		var pts plotter.XYs
		for i := req.NMin; i < req.NMax; i += 1 {
			var exactPts plotter.XYs
			err = s.ExactSolver.Solve(solver.CalculateStepSize(i, req.X0, req.XEnd), req.X0, req.Y0, req.XEnd, solver.DrawerFunc(func(ps solver.Point) error {
				exactPts = append(exactPts, ps.XY())
				return nil
			}))
			if err != nil {
				rest.SendErrorHTML(w, r, http.StatusInternalServerError, err, "can't solve for gtes with exact")
				return
			}

			var solPts plotter.XYs
			err = slvr.Solve(solver.CalculateStepSize(i, req.X0, req.XEnd), req.X0, req.Y0, req.XEnd, solver.DrawerFunc(func(ps solver.Point) error {
				solPts = append(solPts, ps.XY())
				return nil
			}))
			if err != nil {
				rest.SendErrorHTML(w, r, http.StatusInternalServerError, err, fmt.Sprintf("can't solve for gtes with %s", slvr.Name()))
				return
			}

			var mx plotter.XY
			for j := range exactPts {
				y := math.Abs(solPts[j].Y - exactPts[j].Y)
				if j == 0 || y > mx.Y {
					mx = plotter.XY{X: exactPts[j].X, Y: y}
				}
			}
			pts = append(pts, mx)
		}
		gte = append(gte, graph.Line{
			Name:   slvr.Name(),
			Points: pts,
		})
	}
	b, err = s.Plotter.Plot("GTE", gte)
	if err != nil {
		rest.SendErrorHTML(w, r, http.StatusInternalServerError, err, "can't plot the gte graph")
		return
	}
	b64GTEGraph := base64.StdEncoding.EncodeToString(b)

	// plotting solutions graph
	lines = append(lines, graph.Line{Name: "Exact solution", Points: exactPts})
	b, err = s.Plotter.Plot("Solutions", lines)
	if err != nil {
		rest.SendErrorHTML(w, r, http.StatusInternalServerError, err, "can't plot the solutions graph")
		return
	}
	b64SolGraph := base64.StdEncoding.EncodeToString(b)

	// building html template
	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("plot").Parse(plotHTMLTmpl))
	err = tmpl.Execute(buf, plotTmplData{
		X0:           req.X0,
		Y0:           req.Y0,
		XEnd:         req.XEnd,
		N:            req.N,
		SolutionsImg: b64SolGraph,
		GTEImg:       b64GTEGraph,
		LTEImg:       b64LTEGraph,
	})
	if err != nil {
		rest.SendErrorHTML(w, r, http.StatusInternalServerError, err, "can't execute template")
		return
	}

	render.Status(r, http.StatusOK)
	render.HTML(w, r, buf.String())
}

type solveRequest struct {
	X0   float64
	Y0   float64
	XEnd float64
	N    int
	NMin int
	NMax int
}

func readVals(v url.Values) (req solveRequest, err error) {
	var x0, y0, xEnd float64
	var n, nmin, nmax int

	if len(v["x0"]) != 1 || len(v["y0"]) != 1 || len(v["x_end"]) != 1 || len(v["n"]) != 1 || len(v["nmin"]) != 1 || len(v["nmax"]) != 1 {
		return solveRequest{}, errors.New("some fields are empty or contains more or less entries, than needed")
	}
	if err := json.Unmarshal([]byte(v["x0"][0]), &x0); err != nil {
		return solveRequest{}, errors.Wrap(err, "can't read x0")
	}
	if err := json.Unmarshal([]byte(v["y0"][0]), &y0); err != nil {
		return solveRequest{}, errors.Wrap(err, "can't read y0")
	}
	if err := json.Unmarshal([]byte(v["x_end"][0]), &xEnd); err != nil {
		return solveRequest{}, errors.Wrap(err, "can't read xEnd")
	}
	if err := json.Unmarshal([]byte(v["n"][0]), &n); err != nil {
		return solveRequest{}, errors.Wrap(err, "can't read n")
	}
	if err := json.Unmarshal([]byte(v["nmin"][0]), &nmin); err != nil {
		return solveRequest{}, errors.Wrap(err, "can't read nmin")
	}
	if err := json.Unmarshal([]byte(v["nmax"][0]), &nmax); err != nil {
		return solveRequest{}, errors.Wrap(err, "can't read nmax")
	}

	return solveRequest{X0: x0, Y0: y0, XEnd: xEnd, N: n, NMax: nmax, NMin: nmin}, nil
}
