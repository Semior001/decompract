package solver

import (
	"fmt"
	"testing"

	log "github.com/go-pkgz/lgr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSolvers(t *testing.T) {
	log.Setup(log.Debug, log.CallerFile, log.CallerFunc, log.Msec, log.LevelBraces)
	tbl := []struct {
		solver    Interface
		name      string
		points    []Point
		precision float64
	}{
		{
			solver: &Euler{F: func(x, y float64) (float64, error) { return x*x - 2.0*y, nil }},
			name:   "Euler",
			points: []Point{
				{0.0, 1.00000},
				{0.1, 0.80000},
				{0.2, 0.64100},
				{0.3, 0.51680},
				{0.4, 0.42244},
				{0.5, 0.35395},
				{0.6, 0.30816},
				{0.7, 0.28253},
				{0.8, 0.27502},
				{0.9, 0.28402},
				{1.0, 0.30821},
			},
			precision: 0.00001,
		},
		{
			solver: &ImprovedEuler{F: func(x, y float64) (float64, error) { return x*x - 2.0*y, nil }},
			name:   "Improved Euler",
			points: []Point{
				{0.0, 1.000000},
				{0.1, 0.820250},
				{0.2, 0.674755},
				{0.3, 0.559149},
				{0.4, 0.469852},
				{0.5, 0.403929},
				{0.6, 0.358972},
				{0.7, 0.333007},
				{0.8, 0.324416},
				{0.9, 0.331871},
				{1.0, 0.354284},
			},
			precision: 0.000001,
		},
		{
			solver: &RungeKutta{F: func(x, y float64) (float64, error) { return x*x - 2.0*y, nil }},
			name:   "Runge-Kutta",
			points: []Point{
				{0.0, 1.000000},
				{0.1, 0.819051},
				{0.2, 0.672745},
				{0.3, 0.556615},
				{0.4, 0.467004},
				{0.5, 0.400917},
				{0.6, 0.355903},
				{0.7, 0.329955},
				{0.8, 0.321430},
				{0.9, 0.328982},
				{1.0, 0.351509},
			},
			precision: 0.000001,
		},
	}

	for _, entry := range tbl {
		step := 0
		err := entry.solver.Solve(0.1, 0, 1, 1, DrawerFunc(func(ps Point) error {
			assert.InDelta(t, entry.points[step].X, ps.X, entry.precision, fmt.Sprintf("Method: %s, step: %d", entry.name, step))
			assert.InDelta(t, entry.points[step].Y, ps.Y, entry.precision, fmt.Sprintf("Method: %s, step: %d", entry.name, step))
			step++
			return nil
		}))
		require.NoError(t, err)
	}
}
