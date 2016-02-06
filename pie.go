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
	Shift  uint
}

type PieDiagram struct {
	Title  string
	Width  uint
	Height uint
	ShowValues bool
	ValuesShift uint
	Radius uint

	categories []*PieCategory
	total float64
	graphWidth uint
	graphHeight uint
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

	var radius uint
	d.graphWidth = uint(float64(d.Width) * 0.66)

	d.graphHeight = uint(d.Height) - dsMarginTop

	if d.graphWidth > d.graphHeight {
		radius = (d.graphHeight - 2*dsPieMargin)/2
	} else {
		radius = (d.graphWidth - 2*dsPieMargin)/2
	}

	// Calculate max radius depending on max PieCategory.Shift
	var maxShift uint
	for _, cat := range d.categories {
		if cat.Shift > maxShift {
			maxShift = cat.Shift
		}
	}

	radius -= maxShift


	if d.Radius == 0 || d.Radius > radius {
		d.Radius = radius
	}

	if d.ShowValues && d.ValuesShift == 0 {
		d.ValuesShift = d.Radius + dsPieMargin/2
	}


	return
}

func (d *PieDiagram) build(w io.Writer) (err error) {

	if err = d.validate(); err != nil {
		return
	}

	s := svg.New(w)
	s.Start(int(d.Width), int(d.Height))

	// Title
	s.Text(int(d.Width)/2, dsMarginTop/2, d.Title,
		fmt.Sprintf("text-anchor:middle;alignment-baseline:central;font-size:%d;fill:%s",
			dsTitleFontSize, dsTitleFontColor))


	var cx int = int(d.graphWidth/2)
	var cy int = dsMarginTop + int(d.graphHeight/2)

	// Calculate height and start for legend
	var lHeight int = dsLegendMarkSize
	if dsLegendFontSize > dsLegendMarkSize  {
		lHeight = dsLegendFontSize
	}
	lx := d.graphWidth
	ly := dsMarginTop + dsPieMargin

	if len(d.categories) > 1 {

		var lastx int = int(d.Radius)
		var lasty int = 0
		var seg float64 = 0

		for _, cat := range d.categories {

			arc := "0"
			seg = cat.Value / d.total * 360 + seg
			if (cat.Value / d.total * 360) > 180 {
				arc = "1"
			}
			var radseg float64 = math.Pi / 180.0 * seg
			var nextx int = int(math.Cos(radseg) * float64(d.Radius))
			var nexty int = int(math.Sin(radseg) * float64(d.Radius))

			radseg = math.Pi / 180.0 * (seg - (cat.Value / d.total * 360)/2)
			var sx int = int(math.Cos(radseg) * float64(cat.Shift))
			var sy int = int(math.Sin(radseg) * float64(cat.Shift))

			x := cx + sx
			y := cy - sy

			path := fmt.Sprintf("M %d,%d l %d,%d a%d,%d 0 " + arc + ",0 %d,%d z",
				x, y,
				lastx, -lasty,
				d.Radius, d.Radius,
				nextx - lastx,
				-(nexty - lasty))

			s.Path(path, "fill:" + cat.Color)

			if d.ShowValues {
				// Draw value

				var sx int = int(math.Cos(radseg) * float64(d.ValuesShift))
				var sy int = int(math.Sin(radseg) * float64(d.ValuesShift))
				x := cx + sx
				y := cy - sy

				s.Text(x, y, fmt.Sprintf("%.2f", cat.Value),
					fmt.Sprintf("text-anchor:middle;font-size:%d;fill:%s", dsLabelsFontSize, dsLabelsFontColor))
			}

			lastx = nextx
			lasty = nexty

			// Draw legend
			// TODO draw legend in any side
			// TODO do not draw legend if it's do not fit?
			s.Rect(int(lx), int(ly), dsLegendMarkSize, dsLegendMarkSize,
				fmt.Sprintf("fill:%s", cat.Color))
			s.Text(int(lx)+dsLegendMarkSize+5, ly+lHeight/2+dsLegendFontSize/2, cat.Name,
				fmt.Sprintf("font-size:%d;fill:%s", dsLegendFontSize, dsLabelsFontColor))
			ly += lHeight + dsLegendMargin
		}




	} else {
		s.Circle(cx, cy, int(d.Radius), "fill:"+d.categories[0].Color)
		if d.ShowValues {
			// Draw value
			s.Text(cx, cy, fmt.Sprintf("%.2f", d.categories[0].Value),
				fmt.Sprintf("text-anchor:middle;font-size:%d;fill:%s", dsLabelsFontSize, dsLabelsFontColor))
		}
	}

	s.End()

	return
}
