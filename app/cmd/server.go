package cmd

import (
	"math"

	"github.com/Semior001/decompract/app/num/graph"

	"github.com/Semior001/decompract/app/num/service"

	"github.com/Semior001/decompract/app/num/solver"
	"github.com/Semior001/decompract/app/rest/api"
)

// Server runs REST API web server
type Server struct {
	ServiceURL string `long:"service_url" env:"SERVICE_URL" description:"http service url" default:"http://localhost:8080"`
	Port       int    `long:"service_port" env:"SERVICE_PORT" description:"http server port" default:"8080"`

	WebRoot string `long:"web-root" env:"WEB_ROOT" default:"./web" description:"web root directory"`

	CommonOpts
}

// Execute runs http web server
func (s *Server) Execute(_ []string) error {
	//fxy := func(x, y float64) (float64, error) { return x*x - 2.0*y, nil }
	fxy := func(x, y float64) (float64, error) {
		return y*y*math.Exp(x) - 2.0*y, nil
	}

	srv := api.Rest{
		Version: s.Version,
		WebRoot: s.WebRoot,
		NumService: &service.Service{
			Solvers: []solver.Interface{
				&solver.RungeKutta{F: fxy},
				&solver.ImprovedEuler{F: fxy},
				&solver.Euler{F: fxy},
			},
			ExactSolver: &solver.Exact{
				//F: func(x, c float64) (float64, error) { return c*math.Exp(-2*x) + x*x/2 - x/2 + 1/4, nil },
				//C: func(x0, y0 float64) (float64, error) {
				//	return (y0 - (x0*x0)/2.0 - x0/2.0 - 1/4) / (math.Exp(-2 * x0)), nil
				//},
				F: func(x, c float64) (float64, error) { return math.Exp(-x) / (c*math.Exp(x) + 1), nil },
				C: func(x0, y0 float64) (float64, error) { return (math.Exp(-x0) - y0) / (y0 * math.Exp(x0)), nil },
			},
			Plotter: graph.Plotter{},
		},
	}
	srv.Run(s.Port)
	return nil
}
