package solver

import (
	log "github.com/go-pkgz/lgr"
	"github.com/pkg/errors"
)

// Euler method for solving initial value problem for differential equations
type Euler struct {
	F func(x, y float64) (float64, error) // calculator for f(x,y) = y'
}

// Name returns the name of this method
func (*Euler) Name() string {
	return "Euler's method"
}

// Solve the initial value problem with Euler method
func (e *Euler) Solve(stepSize, x0, y0, xEnd float64, dr Drawer) error {
	x := x0
	y := y0
	var f float64
	var err error

	log.Printf("[DEBUG] starting solving the equation with Euler's "+
		"method with stepsz = %.4f, x0 = %.4f, y0 = %.4f, xend = %.4f", stepSize, x0, y0, xEnd)

	for x < xEnd {
		if err := dr.Draw(Point{X: x, Y: y}); err != nil {
			return errors.Wrapf(err, "failed to draw a point (%.4f, %.4f)", x, y)
		}

		if f, err = e.F(x, y); err != nil {
			return errors.Wrapf(err, "failed to calculate f for x=%.4f y=%.4f", x, y)
		}

		// calculating the next x, y values
		y = e.calculateY(y, stepSize*f)
		x += stepSize
	}

	return nil
}

// calculate y value as
// y_{i+1} = y_i + h * f(x_i, y_i)
func (e *Euler) calculateY(yi, hf float64) float64 {
	return yi + hf
}
