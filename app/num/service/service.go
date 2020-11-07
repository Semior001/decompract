// Package num provides methods to solve differential equations in num methods,
// plot graphs of these solutions, calculate errors and plot the graphs of errors
package service

import (
	"github.com/Semior001/decompract/app/num"
	"github.com/Semior001/decompract/app/num/graph"
	"github.com/Semior001/decompract/app/num/solver"
	"github.com/pkg/errors"
	"math"
)

// Service provides methods to operate solvers, plot a graph, calculate errors
type Service struct {
	Plotter     graph.Plotter
	Solvers     []solver.Interface
	ExactSolver solver.Interface
}

// solve returns the lines with the num solutions of the differential equation
// with the given input data, without the exact solution
func (s *Service) solve(stepSize, x0, y0, xEnd float64) ([]num.Line, error) {
	var lines []num.Line
	// solving equation
	for _, slvr := range s.Solvers {
		line, err := slvr.Solve(stepSize, x0, y0, xEnd)
		if err != nil {
			return nil, errors.Wrapf(err, "can't solve")
		}

		lines = append(lines, line)
	}
	return lines, nil
}

// PlotSolutions solves the differential equation by Solvers with the given input data
func (s *Service) PlotSolutions(stepSize, x0, y0, xEnd float64) (plot []byte, err error) {
	lines, err := s.solve(stepSize, x0, y0, xEnd)
	if err != nil {
		return nil, err
	}

	// adding exact solution to the graph
	line, err := s.ExactSolver.Solve(stepSize, x0, y0, xEnd)
	if err != nil {
		return nil, errors.Wrapf(err, "can't solve with exact solution")
	}
	lines = append(lines, line)

	// plotting the solutions
	if plot, err = s.Plotter.Plot("Solutions", lines); err != nil {
		return nil, errors.Wrapf(err, "can't plot graph")
	}

	return plot, nil
}

// PlotLocalErrors plots the graph of truncation errors from solvers related to the exact
// solution
func (s *Service) PlotLocalErrors(stepSize, x0, y0, xEnd float64) (plot []byte, err error) {
	solLines, err := s.solve(stepSize, x0, y0, xEnd)
	if err != nil {
		return nil, err
	}

	// getting the exact solution
	exactLine, err := s.ExactSolver.Solve(stepSize, x0, y0, xEnd)
	if err != nil {
		return nil, errors.Wrap(err, "can't solve with exact solution")
	}

	// calculating and aggregating truncation errors
	var errLines []num.Line
	for _, line := range solLines {
		if len(line.Points) != len(exactLine.Points) {
			return nil, errors.Errorf("number of points are different for exact and %s", line.Name)
		}
		var pts []num.Point
		for i := range exactLine.Points {
			if exactLine.Points[i].X != line.Points[i].X {
				return nil, errors.Errorf("x coord are different for exact and %s at i=%d", line.Name, i)
			}

			// calculating error by Y
			y := math.Abs(line.Points[i].Y - exactLine.Points[i].Y)
			pts = append(pts, num.Point{X: exactLine.Points[i].X, Y: y})
		}
		errLines = append(errLines, num.Line{Name: line.Name, Points: pts})
	}

	// plotting the solutions
	if plot, err = s.Plotter.Plot("LTE", errLines); err != nil {
		return nil, errors.Wrapf(err, "can't plot graph")
	}
	return plot, nil
}
