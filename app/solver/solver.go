package solver

import "fmt"

// Interface describes methods that the solver should implement
// in order to solve the Initial Value problem
type Interface interface {
	Solve(stepSize, x0, y0, xEnd float64, dr Drawer) (err error)
}

// Drawer draws a line by the given set of points
type Drawer interface {
	Draw(ps Point) error
}

// DrawerFunc is an adapter to use ordinary function as Drawer
type DrawerFunc func(ps Point) error

// Draw the graph by the given set of points and name
func (f DrawerFunc) Draw(ps Point) error { return f(ps) }

// Point describes a particular point on a plane
type Point struct {
	X float64
	Y float64
}

// String implements fmt.Stringer to properly print points
func (p Point) String() string {
	return fmt.Sprintf("(%.4f, %.4f)", p.X, p.Y)
}

// CalculateStepSize from the given number of steps
func CalculateStepSize(n int, x0, x float64) float64 {
	return (x - x0) / float64(n)
}
