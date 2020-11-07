package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rakyll/statik/fs"

	"github.com/Semior001/decompract/app/num/solver"

	"github.com/Semior001/decompract/app/num/service"
	// adding statik files, for serving form
	_ "github.com/Semior001/decompract/app/statik"

	"github.com/Semior001/decompract/app/num"

	"github.com/pkg/errors"

	"github.com/go-chi/render"

	"github.com/Semior001/decompract/app/rest"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/httprate"
	log "github.com/go-pkgz/lgr"
	R "github.com/go-pkgz/rest"

	"github.com/Knetic/govaluate"
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
    <p>f(x,y) = {{.Fxy}}; y(x,c) = {{.Yxc}}; C(x<sub>0</sub>,y<sub>0</sub>) = {{.Cx0y0}}</p>
    <p>x<sub>0</sub> = {{printf "%.4f" .X0}}; y<sub>0</sub> = {{printf "%.4f" .Y0}}; X = {{printf "%.4f" .XEnd}}; N = {{.N}}; N<sub>min</sub> = {{.NMin}}; N<sub>max</sub> = {{.NMax}}</p>
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
	NMin         int
	NMax         int
	SolutionsImg string
	LTEImg       string
	GTEImg       string
	Fxy          string
	Yxc          string
	Cx0y0        string
}

// Rest defines a simple web server for routing to calendar REST api methods
type Rest struct {
	Version string
	WebRoot string

	NumService *service.Service

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
	var webFS http.Handler

	statikFS, err := fs.New()
	if err != nil {
		log.Printf("[DEBUG] no embedded assets loaded, %s", err)
		log.Printf("[INFO] run file server for %s, path %s", root, path)
		webFS = http.FileServer(root)
	} else {
		log.Printf("[INFO] run file server for %s, embedded", root)
		webFS = http.FileServer(statikFS)
	}

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

// POST / - plot graphs according to the given parameters
func (s *Rest) plotGraphsCtrl(w http.ResponseWriter, r *http.Request) {
	// reading form
	req, err := readVals(r)
	if err != nil {
		rest.SendErrorHTML(w, r, http.StatusForbidden, err, "failed to read request values")
		return
	}

	// if functions are specified, prepare them
	if req.fxy != "" && req.yxc != "" && req.c != "" {
		funcs, err := prepareFuncs(req.fxy, req.yxc, req.c)
		if err != nil {
			rest.SendErrorHTML(w, r, http.StatusBadRequest, err, "failed to parse functions")
			return
		}

		// initializing services
		s.NumService = &service.Service{
			Plotter: s.NumService.Plotter,
			Solvers: []solver.Interface{
				&solver.RungeKutta{F: funcs.fxy},
				&solver.ImprovedEuler{F: funcs.fxy},
				&solver.Euler{F: funcs.fxy},
			},
			ExactSolver: &solver.Exact{F: funcs.yxc, C: funcs.cx0y0},
		}
	}

	// encoding solutions plot
	bSols, err := s.NumService.PlotSolutions(num.CalculateStepSize(req.N, req.X0, req.XEnd), req.X0, req.Y0, req.XEnd)
	if err != nil {
		rest.SendErrorHTML(w, r, http.StatusInternalServerError, err, "failed to plot solutions")
		return
	}

	// encoding lte plot
	bLTEs, err := s.NumService.PlotLocalErrors(num.CalculateStepSize(req.N, req.X0, req.XEnd), req.X0, req.Y0, req.XEnd)
	if err != nil {
		rest.SendErrorHTML(w, r, http.StatusInternalServerError, err, "failed to plot lte")
		return
	}

	// encoding gte plot
	bGTEs, err := s.NumService.PlotGlobalErrors(req.NMin, req.NMax, req.X0, req.Y0, req.XEnd)
	if err != nil {
		rest.SendErrorHTML(w, r, http.StatusInternalServerError, err, "failed to plot gte")
		return
	}

	// building html template
	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("plot").Parse(plotHTMLTmpl))
	err = tmpl.Execute(buf, plotTmplData{
		X0:           req.X0,
		Y0:           req.Y0,
		XEnd:         req.XEnd,
		N:            req.N,
		NMin:         req.NMin,
		NMax:         req.NMax,
		SolutionsImg: base64.StdEncoding.EncodeToString(bSols),
		LTEImg:       base64.StdEncoding.EncodeToString(bLTEs),
		GTEImg:       base64.StdEncoding.EncodeToString(bGTEs),
		Fxy:          req.fxy,
		Yxc:          req.yxc,
		Cx0y0:        req.c,
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
	fxy  string
	yxc  string
	c    string
}

func readVals(r *http.Request) (req solveRequest, err error) {
	if err := r.ParseForm(); err != nil {
		return solveRequest{}, errors.Wrap(err, "failed to parse form data")
	}

	var x0, y0, xEnd float64
	var n, nmin, nmax int

	if len(r.Form["x0"]) != 1 || len(r.Form["y0"]) != 1 ||
		len(r.Form["x_end"]) != 1 || len(r.Form["n"]) != 1 ||
		len(r.Form["nmin"]) != 1 || len(r.Form["nmax"]) != 1 {
		return solveRequest{}, errors.New("some fields are empty or contains more or less entries, than needed")
	}
	if err := json.Unmarshal([]byte(r.Form["x0"][0]), &x0); err != nil {
		return solveRequest{}, errors.Wrap(err, "can't read x0")
	}
	if err := json.Unmarshal([]byte(r.Form["y0"][0]), &y0); err != nil {
		return solveRequest{}, errors.Wrap(err, "can't read y0")
	}
	if err := json.Unmarshal([]byte(r.Form["x_end"][0]), &xEnd); err != nil {
		return solveRequest{}, errors.Wrap(err, "can't read xEnd")
	}
	if err := json.Unmarshal([]byte(r.Form["n"][0]), &n); err != nil {
		return solveRequest{}, errors.Wrap(err, "can't read n")
	}
	if err := json.Unmarshal([]byte(r.Form["nmin"][0]), &nmin); err != nil {
		return solveRequest{}, errors.Wrap(err, "can't read nmin")
	}
	if err := json.Unmarshal([]byte(r.Form["nmax"][0]), &nmax); err != nil {
		return solveRequest{}, errors.Wrap(err, "can't read nmax")
	}

	return solveRequest{
		X0:   x0,
		Y0:   y0,
		XEnd: xEnd,
		N:    n,
		NMax: nmax,
		NMin: nmin,
		fxy:  r.Form["fxy"][0],
		yxc:  r.Form["yxc"][0],
		c:    r.Form["c"][0],
	}, nil
}

type parsedFuncs struct {
	fxy   func(x, y float64) (float64, error)
	yxc   func(x, c float64) (float64, error)
	cx0y0 func(x0, y0 float64) (float64, error)
}

// prepareFuncs parses the string expressions and prepares the functions for the future evaluation
func prepareFuncs(fxyStr, yxcStr, cStr string) (parsedFuncs, error) {
	funcs := map[string]govaluate.ExpressionFunction{
		"exp": govaluate.ExpressionFunction(func(args ...interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, errors.New("exponent takes only 1 argument")
			}
			p, ok := args[0].(float64)
			if !ok {
				return nil, errors.New("argument is not of type float64")
			}
			return math.Exp(p), nil
		}),
		"ln": govaluate.ExpressionFunction(func(args ...interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, errors.New("ln takes only 1 argument")
			}
			p, ok := args[0].(float64)
			if !ok {
				return nil, errors.New("argument is not of type float64")
			}
			return math.Log(p), nil
		}),
		"tan": govaluate.ExpressionFunction(func(args ...interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, errors.New("tan takes only 1 argument")
			}
			p, ok := args[0].(float64)
			if !ok {
				return nil, errors.New("argument is not of type float64")
			}
			return math.Tan(p), nil
		}),
		"cos": govaluate.ExpressionFunction(func(args ...interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, errors.New("cos takes only 1 argument")
			}
			p, ok := args[0].(float64)
			if !ok {
				return nil, errors.New("argument is not of type float64")
			}
			return math.Cos(p), nil
		}),
		"pi": govaluate.ExpressionFunction(func(args ...interface{}) (interface{}, error) {
			if len(args) > 0 {
				return nil, errors.New("pi does not require any arguments")
			}
			return math.Pi, nil
		}),
		"sin": govaluate.ExpressionFunction(func(args ...interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, errors.New("sin takes only 1 argument")
			}
			p, ok := args[0].(float64)
			if !ok {
				return nil, errors.New("argument is not of type float64")
			}
			return math.Sin(p), nil
		}),
	}

	fxyExpr, err := govaluate.NewEvaluableExpressionWithFunctions(fxyStr, funcs)
	if err != nil {
		return parsedFuncs{}, errors.Wrap(err, "can't parse f(x,y)")
	}
	fxy := func(x, y float64) (float64, error) {
		params := map[string]interface{}{"x": x, "y": y}
		resExpr, err := fxyExpr.Evaluate(params)
		if err != nil {
			return 0, errors.Wrap(err, "failed to evaluate expression")
		}
		res, ok := resExpr.(float64)
		if !ok {
			return 0, errors.Wrap(err, "result is not float64")
		}
		return res, nil
	}

	yxcExpr, err := govaluate.NewEvaluableExpressionWithFunctions(yxcStr, funcs)
	if err != nil {
		return parsedFuncs{}, errors.Wrap(err, "can't parse y(x,c)")
	}

	cExpr, err := govaluate.NewEvaluableExpressionWithFunctions(cStr, funcs)
	if err != nil {
		return parsedFuncs{}, errors.Wrap(err, "can't parse c(x0,y0)")
	}

	yxc := func(x, c float64) (float64, error) {
		params := map[string]interface{}{"x": x, "c": c}
		resExpr, err := yxcExpr.Evaluate(params)
		if err != nil {
			return 0, errors.Wrap(err, "failed to evaluate expression")
		}
		res, ok := resExpr.(float64)
		if !ok {
			return 0, errors.Wrap(err, "result is not float64")
		}
		return res, nil
	}

	cx0y0 := func(x0, y0 float64) (float64, error) {
		params := map[string]interface{}{"x0": x0, "y0": y0}
		resExpr, err := cExpr.Evaluate(params)
		if err != nil {
			return 0, errors.Wrap(err, "failed to evaluate expression")
		}
		res, ok := resExpr.(float64)
		if !ok {
			return 0, errors.Wrap(err, "result is not float64")
		}
		return res, nil
	}

	return parsedFuncs{fxy: fxy, yxc: yxc, cx0y0: cx0y0}, nil
}
