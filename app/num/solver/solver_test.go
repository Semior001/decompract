package solver

import (
	"math"
	"testing"

	"github.com/Semior001/decompract/app/num"

	log "github.com/go-pkgz/lgr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSolvers(t *testing.T) {
	log.Setup(log.Debug, log.CallerFile, log.CallerFunc, log.Msec, log.LevelBraces)
	tbl := []struct {
		solver    Interface
		name      string
		points    []num.Point
		precision float64
	}{
		{
			solver: &Euler{F: func(x, y float64) (float64, error) { return x*x - 2.0*y, nil }},
			name:   "Euler's method",
			points: []num.Point{
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
			name:   "Improved Euler's method",
			points: []num.Point{
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
			name:   "Runge-Kutta's method",
			points: []num.Point{
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
		line, err := entry.solver.Solve(0.1, 0, 1, 1)
		require.NoError(t, err)
		assert.Equal(t, len(entry.points), len(line.Points), "len are different for method %s", entry.name)
		assert.Equal(t, entry.name, line.Name, "names, method: %s", entry.name)

		for i := range line.Points {
			assert.InDelta(t, entry.points[i].X, line.Points[i].X, entry.precision, "method: %s, step: %d", entry.name, i)
		}
	}
}
func TestExact_Solve(t *testing.T) {
	e := &Exact{
		F: func(x, c float64) (float64, error) { return math.Exp(-x) / (c*math.Exp(x) + 1), nil },
		C: func(x0, y0 float64) (float64, error) { return (math.Exp(-x0) - y0) / (y0 * math.Exp(x0)), nil }, // 2926.3598370085842
	}
	points := []num.Point{
		{-4.0, 1.00000000},
		{-3.5, 0.37054986},
		{-3.0, 0.13692051},
		{-2.5, 0.05050571},
		{-2.0, 0.01861037},
		{-1.5, 0.00685316},
		{-1.0, 0.00252266},
		{-0.5, 0.00092837},
		{+0.0, 0.00034160},
		{+0.5, 0.00012569},
		{+1.0, 0.00004624},
		{+1.5, 0.00001701},
		{+2.0, 0.00000626},
		{+2.5, 0.00000230},
		{+3.0, 0.00000085},
		{+3.5, 0.00000031},
		{+4.0, 0.00000011},
	}

	line, err := e.Solve(0.5, -4, 1, 4)
	require.NoError(t, err)
	assert.Equal(t, len(points), len(line.Points), "len are different")
	assert.Equal(t, "Exact solution", line.Name, "names")

	for i := range line.Points {
		assert.InDelta(t, points[i].X, line.Points[i].X, 0.00000001, "step: %d", i)
	}
}

func TestPoint_String(t *testing.T) {
	assert.Equal(t, "(0.0003, 0.1235)", num.Point{0.0003, 0.123456789}.String())
}

func TestCalculateStepSize(t *testing.T) {
	assert.InDelta(t, 0.26667, num.CalculateStepSize(30, -4.0, 4.0), 0.00001)
}
