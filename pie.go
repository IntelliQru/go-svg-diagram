package svgd

import (
	"errors"
	"fmt"
	"github.com/ajstarks/svgo"
	"io"
	"math/rand"
	"math"
)

type PieCategory struct {
	Name   string
	Color  string
	Value  float64
}

type PieDiagram struct {
	Title  string
	Width  int
	Height int
	ShowValues bool

	categories []*PieCategory
	total float64
}

func (d *PieDiagram) NewCategory(name string) (cat *PieCategory) {
	n := new(PieCategory)
	n.Name = name
	d.categories = append(d.categories, n)

	return n
}

func (d *PieDiagram) validate() (err error) {

	if len(d.categories) == 0 {
		err = errors.New("Error: Nothing to build, categories are empty")
	}

	d.total = 0
	for _, cat := range d.categories {
		// Generate random color if it's doesn't set
		if cat.Color == "" {
			cat.Color = fmt.Sprintf("#%x%x%x", rand.Intn(255), rand.Intn(255), rand.Intn(255))
		}
		d.total += cat.Value
	}

	if d.total == 0 {
		err = errors.New("Error: Total is zero")
	}

	return
}

func (d *PieDiagram) build(w io.Writer) (err error) {

	if err = d.validate(); err != nil {
		return
	}

	s := svg.New(w)
	s.Start(d.Width, d.Height)

	// Title
	s.Text(d.Width/2, dsMarginTop/2, d.Title,
		fmt.Sprintf("text-anchor:middle;alignment-baseline:central;font-size:%d;fill:%s",
			dsTitleFontSize, dsTitleFontColor))


	var radius int
	var graphWidth int = d.Width - dsMarginLeft - dsMarginRight
	var graphHeight int = d.Height - dsMarginBottom - dsMarginTop
	if graphWidth > graphHeight {
		radius = (graphHeight - 2*dsPieMargin)/2
	} else {
		radius = (graphWidth - 2*dsPieMargin)/2
	}

	var cx int = dsMarginLeft + (d.Width - dsMarginLeft - dsMarginRight)/2
	var cy int = d.Height - dsMarginBottom - (d.Height - dsMarginBottom - dsMarginTop)/2

	if len(d.categories) > 1 {

		var lastx int = radius
		var lasty int = 0
		var seg float64 = 0

		for _, cat := range d.categories {

			arc := "0"
			seg = cat.Value / d.total * 360 + seg
			if (cat.Value / d.total * 360) > 180 {
				arc = "1"
			}
			var radseg float64 = math.Pi / 180.0 * seg
			var nextx int = int(math.Cos(radseg) * float64(radius))
			var nexty int = int(math.Sin(radseg) * float64(radius))

			path := fmt.Sprintf("M %d,%d l %d,%d a%d,%d 0 " + arc + ",0 %d,%d z",
				cx, cy,
				lastx, -lasty,
				radius, radius,
				nextx - lastx,
				-(nexty - lasty))

			s.Path(path, "fill:" + cat.Color)

			lastx = nextx
			lasty = nexty
		}
	} else {
		s.Circle(cx, cy, radius, "fill:"+d.categories[0].Color)
	}

	s.End()

	return
}
