package solver

import (
	"github.com/Semior001/decompract/app/num"
	log "github.com/go-pkgz/lgr"
	"github.com/pkg/errors"
)

// Euler method for solving initial value problem for differential equations
type Euler struct {
	F func(x, y float64) (float64, error) // calculator for f(x,y) = y'
}

// Solve the initial value problem with Euler method
func (e *Euler) Solve(stepSize, x0, y0, xEnd float64) (num.Line, error) {
	x := x0
	y := y0
	var f float64
	var err error

	log.Printf("[DEBUG] starting solving the equation with Euler's "+
		"method with stepsz = %.4f, x0 = %.4f, y0 = %.4f, xend = %.4f", stepSize, x0, y0, xEnd)

	var pts []num.Point
	for x <= xEnd {
		pts = append(pts, num.Point{X: x, Y: y})

		if f, err = e.F(x, y); err != nil {
			return num.Line{}, errors.Wrapf(err, "failed to calculate f for x=%.4f y=%.4f", x, y)
		}

		// calculating the next x, y values
		y = e.calculateY(y, stepSize*f)
		x += stepSize
	}

	return num.Line{Name: "Euler's method", Points: pts}, nil
}

// calculate y value as
// y_{i+1} = y_i + h * f(x_i, y_i)
func (e *Euler) calculateY(yi, hf float64) float64 {
	return yi + hf
}
