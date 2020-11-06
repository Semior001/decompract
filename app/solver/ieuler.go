package solver

import (
	log "github.com/go-pkgz/lgr"
	"github.com/pkg/errors"
)

// ImprovedEuler method for solving initial value problem for differential equations
type ImprovedEuler struct {
	F func(x, y float64) (float64, error) // calculator for f(x,y) = y'
}

// Name returns the name of this method
func (*ImprovedEuler) Name() string {
	return "Improved Euler's method"
}

// Solve the differential equations with the given initial data
func (i *ImprovedEuler) Solve(stepSize, x0, y0, xEnd float64, dr Drawer) error {
	x := x0
	y := y0

	log.Printf("[DEBUG] starting solving the equation with Improved Euler's "+
		"method with stepsz = %.4f, x0 = %.4f, y0 = %.4f, xend = %.4f", stepSize, x0, y0, xEnd)

	for x < xEnd {
		if err := dr.Draw(Point{X: x, Y: y}); err != nil {
			return errors.Wrapf(err, "failed to draw a point (%.4f, %.4f)", x, y)
		}

		dy, err := i.calculateDeltaY(stepSize, x, y)
		if err != nil {
			return errors.Wrap(err, "failed to calculate delta y")
		}
		y = y + dy
		x += stepSize
	}

	return nil
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
