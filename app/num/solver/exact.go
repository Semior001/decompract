package solver

import (
	"github.com/Semior001/decompract/app/num"
	"github.com/pkg/errors"
)

// Exact solver just draws the exact solution directly,
// without applying any specific algorithm
type Exact struct {
	// F is a function y(x) = f(x), the solution for the initial value problem,
	// requires a constant, that is calculated with initial values
	F func(x, c float64) (float64, error)
	// C calculates the constant for the F
	C func(x0, y0 float64) (float64, error)
}

// Solve just plots the graph, without applying any algorithm
func (e *Exact) Solve(stepSize, x0, y0, xEnd float64) (num.Line, error) {
	x := x0
	y := y0
	c, err := e.C(x0, y0)
	if err != nil {
		return num.Line{}, errors.Wrapf(err, "failed to calculate constant for x0=%.4f, y0=%.4f", x0, y0)
	}

	var pts []num.Point
	for x <= xEnd {
		pts = append(pts, num.Point{X: x, Y: y})
		x += stepSize
		if y, err = e.F(x, c); err != nil {
			return num.Line{}, errors.Wrapf(err, "failed to calculate y for x=%.4f, c=%.4f", x, c)
		}
	}

	return num.Line{Name: "Exact solution", Points: pts}, nil
}
