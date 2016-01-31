package svgd

import (
	"errors"
	"fmt"
	"github.com/ajstarks/svgo"
	"io"
	"math"
	"math/rand"
)

type BarCategory struct {
	Name   string
	Color  string
	values []float64
}

func (lc *BarCategory) SetValues(vals []float64) {

	lc.values = make([]float64, len(vals))
	copy(lc.values, vals)
}

type BarDiagram struct {
	Title  string
	Width  int
	Height int
	Grid   bool

	MinValue float64
	MaxValue float64
	Step     float64

	categories []*BarCategory
	labels     []string
}

func (d *BarDiagram) NewCategory(name string) (cat *BarCategory) {
	n := new(BarCategory)

	n.values = make([]float64, 0)
	n.Name = name
	d.categories = append(d.categories, n)

	return n
}

func (d *BarDiagram) SetLabels(labels []string) {

	d.labels = make([]string, len(labels))
	copy(d.labels, labels)
}

func (d *BarDiagram) validate() (err error) {

	if d.Step <= 0 {
		err = errors.New("Error: Step must be greater than zero")
	}

	if len(d.categories) == 0 {
		err = errors.New("Error: Nothing to build, categories are empty")
	}

	for _, cat := range d.categories {
		for iVal := 0; iVal < len(cat.values); iVal++ {
			if d.MaxValue < cat.values[iVal] {
				d.MaxValue = cat.values[iVal]
			}
			if d.MinValue > cat.values[iVal] {
				d.MinValue = cat.values[iVal]
			}
		}

		// Generate random color if it's doesn't set
		if cat.Color == "" {
			cat.Color = fmt.Sprintf("#%x%x%x", rand.Intn(255), rand.Intn(255), rand.Intn(255))
		}
	}

	if d.MinValue == d.MaxValue {
		err = errors.New("Error: MaxValue value must be greater than MinValue")
	}

	return
}

func (d *BarDiagram) build(w io.Writer) (err error) {

	if err = d.validate(); err != nil {
		return
	}

	s := svg.New(w)
	s.Start(d.Width, d.Height)

	// Title
	s.Text(d.Width/2, dsMarginTop/2, d.Title,
		fmt.Sprintf("text-anchor:middle;alignment-baseline:central;font-size:%d;fill:%s",
			dsTitleFontSize, dsTitleFontColor))

	// Draw X and Y axis
	s.Line(dsMarginLeft, d.Height-dsMarginBottom, d.Width-dsMarginRight, d.Height-dsMarginBottom,
		fmt.Sprintf("stroke-width:%d;stroke:%s;", dsAxisLineWidth, dsAxisLineColor))
	s.Line(dsMarginLeft, d.Height-dsMarginBottom, dsMarginLeft, dsMarginTop,
		fmt.Sprintf("stroke-width:%d;stroke:%s;", dsAxisLineWidth, dsAxisLineColor))

	// Write labels
	lenLabels := len(d.labels)
	segmentWidth := (d.Width - dsMarginLeft - dsMarginRight) / lenLabels
	left := dsMarginLeft + segmentWidth/2
	s.Group(fmt.Sprintf("text-anchor:middle;font-size:%d;fill:%s", dsLabelsFontSize, dsLabelsFontColor))
	for i := 0; i < lenLabels; i++ {
		s.Text(left, d.Height-dsMarginBottom+dsLabelsMargin, d.labels[i])
		left += segmentWidth
	}
	s.Gend()

	// Round minimum value to nearest multiple of step
	rem := math.Abs(math.Remainder(d.MinValue, d.Step))
	if rem > 0 {
		d.MinValue -= rem
	}
	rem = math.Abs(math.Remainder(d.MaxValue, d.Step))
	if rem > 0 {
		d.MaxValue += rem
	}

	// Calculate dimensions
	var graphHeight int = d.Height - dsMarginBottom - dsMarginTop
	var valSegment float64 = d.MaxValue - d.MinValue
	var stepsCount int = int(valSegment/d.Step+0.5) + 1
	var stepHeight int = graphHeight / (stepsCount - 1)

	// Write Y values
	textValue := d.MinValue
	top := d.Height - dsMarginBottom

	s.Group(fmt.Sprintf("text-anchor:end;font-size:%d;fill:%s",
		dsLabelsFontSize, dsLabelsFontColor))
	for i := 0; i < stepsCount; i++ {
		s.Text(dsMarginLeft-dsValuesMargin, top, fmt.Sprintf("%.2f", textValue), "alignment-baseline:central")
		textValue += d.Step
		top -= stepHeight
	}
	s.Gend()

	// Drawing grid
	if d.Grid {

		s.Group("stroke-width:1;stroke:lightgray")

		// Horizontal grid
		top = d.Height - dsMarginBottom - stepHeight
		for i := 1; i < stepsCount; i++ {
			s.Line(dsMarginLeft, top, d.Width-dsMarginRight, top)
			top -= stepHeight
		}

		s.Gend()
	}

	var barsCount int = len(d.categories)
	var barWidth int = segmentWidth/barsCount - dsBarMargin*2

	for i := 0; i < lenLabels; i++ {

		var multiplier float64 = float64(stepHeight) / d.Step
		x := i*segmentWidth + dsMarginLeft + dsBarMargin

		for c := 0; c < barsCount; c++ {
			if len(d.categories[c].values) > i {

				var pointValue float64 = d.categories[c].values[i] - d.MinValue
				var stepsInPointValue int = int(pointValue / d.Step)
				var remain int = int((pointValue - float64(stepsInPointValue)*d.Step) * multiplier)

				var barHeight int = int(pointValue/d.Step)*stepHeight + remain
				y := d.Height - dsMarginBottom - barHeight

				s.Rect(x, y, barWidth, barHeight, fmt.Sprintf("fill:%s", d.categories[c].Color))

				// Draw value
				s.Text(x+barWidth/2, y-dsBarMargin, fmt.Sprintf("%.2f", d.categories[c].values[i]),
					fmt.Sprintf("text-anchor:middle;font-size:%d;fill:%s", dsLabelsFontSize, dsLabelsFontColor))
			}
			x += barWidth + dsBarMargin
		}

	}

	// Calculate height and start for legend
	lHeight := (dsMarginBottom - dsLabelsMargin) / (len(d.categories) + 1)
	lTop := d.Height - dsMarginBottom + dsLabelsMargin + lHeight/2

	for _, cat := range d.categories {
		// Draw legend
		// TODO draw legend in any side
		// TODO do not draw legend if it's do not fit?
		s.Rect(d.Width/2, lTop+lHeight/2-dsLegendMarkSize/2, dsLegendMarkSize, dsLegendMarkSize,
			fmt.Sprintf("fill:%s", cat.Color))
		s.Text(d.Width/2+dsLegendMarkSize+5, lTop+lHeight/2, cat.Name,
			fmt.Sprintf("alignment-baseline:middle;font-size:%d;fill:%s", dsLegendFontSize, dsLabelsFontColor))
		lTop += lHeight
	}

	s.End()

	return
}
