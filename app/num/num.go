package num

import "fmt"

// Line describes a particular line on a plot
type Line struct {
	Name   string
	Points []Point
}

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
