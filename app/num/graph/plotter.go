package graph

import (
	"bytes"

	"github.com/Semior001/decompract/app/num"
	"gonum.org/v1/plot/plotter"

	"gonum.org/v1/plot/plotutil"

	"gonum.org/v1/plot/vg"

	"github.com/pkg/errors"
	"gonum.org/v1/plot"
)

const w = 10 * vg.Inch
const h = w

// Plotter plots the image and saves it in temporary directory
type Plotter struct{}

// Plot the set of lines and get the reader of the result plot
func (pl *Plotter) Plot(title string, lines []num.Line) ([]byte, error) {
	p, err := plot.New()
	if err != nil {
		return nil, errors.Wrap(err, "can't create new plot")
	}

	p.Title.Text = title
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"

	// combining lines in one slice to satisfy idiotic plotutil's interface
	var ls []interface{}
	for _, line := range lines {
		ls = append(ls, line.Name, ptsToXYs(line.Points))
	}

	if err = plotutil.AddLinePoints(p, ls...); err != nil {
		return nil, errors.Wrapf(err, "can't add lines to plot %s", title)
	}

	b := &bytes.Buffer{}

	wt, err := p.WriterTo(w, h, "png")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to instantiate writer for the plot %s", title)
	}

	if _, err := wt.WriteTo(b); err != nil {
		return nil, errors.Wrapf(err, "failed to write plot to buffer for %s", title)
	}

	return b.Bytes(), nil
}

// ptsToXYs converts the service-layer points to plotter's interpretation
func ptsToXYs(pts []num.Point) plotter.XYs {
	var res plotter.XYs
	for _, p := range pts {
		res = append(res, plotter.XY{X: p.X, Y: p.Y})
	}
	return res
}
