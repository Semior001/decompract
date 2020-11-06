package solver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	log "github.com/go-pkgz/lgr"
)

func TestImprovedEuler_Solve(t *testing.T) {
	ie := ImprovedEuler{F: func(x, y float64) (float64, error) { return x*x - 2.0*y, nil }}

	log.Setup(log.Debug, log.CallerFile, log.CallerFunc, log.Msec, log.LevelBraces)
	expected := []Point{
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
	}

	step := 0
	err := ie.Solve(0.1, 0, 1, 1, DrawerFunc(func(ps Point) error {
		assert.InDelta(t, expected[step].X, ps.X, 0.00001)
		assert.InDelta(t, expected[step].Y, ps.Y, 0.00001)
		step++
		return nil
	}))
	require.NoError(t, err)
}
