package solver

import (
	"github.com/Semior001/decompract/app/num"
)

// Interface describes methods that the solver should implement
// in order to solve the Initial Value problem
type Interface interface {
	Solve(stepSize, x0, y0, xEnd float64) (line num.Line, err error)
}
