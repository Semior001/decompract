package solver

import (
	"github.com/Semior001/decompract/app/num"
	log "github.com/go-pkgz/lgr"
	"github.com/pkg/errors"
)

// ImprovedEuler method for solving initial value problem for differential equations
type ImprovedEuler struct {
	F func(x, y float64) (float64, error) // calculator for f(x,y) = y'
}

// Solve the differential equations with the given initial data
func (i *ImprovedEuler) Solve(stepSize, x0, y0, xEnd float64) (num.Line, error) {
	x := x0
	y := y0

	log.Printf("[DEBUG] starting solving the equation with Improved Euler's "+
		"method with stepsz = %.4f, x0 = %.4f, y0 = %.4f, xend = %.4f", stepSize, x0, y0, xEnd)

	var pts []num.Point
	for x <= xEnd {
		pts = append(pts, num.Point{X: x, Y: y})

		dy, err := i.calculateDeltaY(stepSize, x, y)
		if err != nil {
			return num.Line{}, errors.Wrap(err, "failed to calculate delta y")
		}
		y = y + dy
		x += stepSize
	}

	return num.Line{Name: "Improved Euler's method", Points: pts}, nil
}

// calculateDeltaY calculates:
// \delta{y_i} = h*f(x_i + h/2, y_i + f(x_i, y_i) * h/2)
func (i *ImprovedEuler) calculateDeltaY(stepsz, xi, yi float64) (float64, error) {
	fxiyi, err := i.F(xi, yi)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to calculate f(%4.f, %.4f)", xi, yi)
	}
	f, err := i.F(xi+stepsz/2.0, yi+(fxiyi/2.0)*stepsz)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to calculate complex f for h=%.4f, xi=%.4f, yi=%4.f", stepsz, xi, yi)
	}
	return stepsz * f, nil
}
