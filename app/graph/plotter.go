package graph

import (
	"bytes"

	"gonum.org/v1/plot/plotutil"

	"gonum.org/v1/plot/vg"

	"github.com/pkg/errors"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
)

// Plotter plots the image and saves it in temporary directory
type Plotter struct {
	DirPath string
}

// Plot the set of lines and get the reader of the result plot
func (pl *Plotter) Plot(title string, lines []Line) ([]byte, error) {
	p, err := plot.New()
	if err != nil {
		return nil, errors.Wrap(err, "can't create new plot")
	}

	p.Title.Text = title
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"

	var ls []interface{}

	for _, line := range lines {
		ls = append(ls, line.Name, line.Points)
	}

	if err = plotutil.AddLinePoints(p, ls...); err != nil {
		return nil, errors.Wrapf(err, "can't add lines to plot %s", title)
	}

	b := &bytes.Buffer{}

	wt, err := p.WriterTo(16*vg.Inch, 16*vg.Inch, "png")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to instantiate writer for the plot %s", title)
	}

	if _, err := wt.WriteTo(b); err != nil {
		return nil, errors.Wrapf(err, "failed to write plot to buffer for %s", title)
	}

	return b.Bytes(), nil
}

// Line describes a particular line on a plot
type Line struct {
	Name   string
	Points plotter.XYs
}
