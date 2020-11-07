package solver

import (
	"github.com/Semior001/decompract/app/num"
	log "github.com/go-pkgz/lgr"
	"github.com/pkg/errors"
)

// RungeKutta  method for solving initial value problem for differential equations
type RungeKutta struct {
	F func(x, y float64) (float64, error) // calculator for f(x,y) = y'
}

// Solve the differential equation with the given initial values
func (r *RungeKutta) Solve(stepSize, x0, y0, xEnd float64) (num.Line, error) {
	x := x0
	y := y0
	var k1, k2, k3, k4 float64
	var err error

	log.Printf("[DEBUG] starting solving the equation with Runge-Kutta's "+
		"method with stepsz = %.4f, x0 = %.4f, y0 = %.4f, xend = %.4f", stepSize, x0, y0, xEnd)

	var pts []num.Point

	for x <= xEnd {
		pts = append(pts, num.Point{X: x, Y: y})

		if k1, err = r.F(x, y); err != nil {
			return num.Line{}, errors.Wrapf(err, "failed to calculate k1 for x=%.4f y=%.4f", x, y)
		}

		if k2, err = r.F(x+stepSize/2.0, y+(stepSize/2.0)*k1); err != nil {
			return num.Line{}, errors.Wrapf(err, "failed to calculate k2 for h=%.4f, x=%4.f, y=%.4f, k1=%.4f", stepSize, x, y, k1)
		}

		if k3, err = r.F(x+stepSize/2.0, y+(stepSize/2.0)*k2); err != nil {
			return num.Line{}, errors.Wrapf(err, "failed to calculate k3 for h=%.4f, x=%4.f, y=%.4f, k2=%.4f", stepSize, x, y, k2)
		}

		if k4, err = r.F(x+stepSize, y+stepSize*k3); err != nil {
			return num.Line{}, errors.Wrapf(err, "failed to calculate k4 for h=%.4f, x=%4.f, y=%.4f, k3=%.4f", stepSize, x, y, k3)
		}

		deltaY := stepSize / 6.0 * (k1 + 2*k2 + 2*k3 + k4)

		y = y + deltaY
		x += stepSize
	}

	return num.Line{Name: "Runge-Kutta's method", Points: pts}, nil
}
