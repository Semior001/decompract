package solver

import (
	"testing"

	"github.com/stretchr/testify/assert"

	log "github.com/go-pkgz/lgr"

	"github.com/stretchr/testify/require"
)

func TestEuler_Solve(t *testing.T) {
	e := Euler{F: func(x, y float64) (float64, error) { return x*x - 2.0*y, nil }}

	log.Setup(log.Debug, log.CallerFile, log.CallerFunc, log.Msec, log.LevelBraces)
	expected := []Point{
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
	}

	step := 0
	err := e.Solve(0.1, 0, 1, 1, DrawerFunc(func(ps Point) error {
		assert.InDelta(t, expected[step].X, ps.X, 0.00001)
		assert.InDelta(t, expected[step].Y, ps.Y, 0.00001)
		step++
		return nil
	}))
	require.NoError(t, err)
}
