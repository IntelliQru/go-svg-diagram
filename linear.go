package svgd
import (
	"io"
	"errors"
	"fmt"
	"github.com/ajstarks/svgo"
	"math"
)


type LinearCategory struct {
	Color  string
	LineWidth int
	values []float64

}

func (lc *LinearCategory) SetValues(vals []float64)  {

	lc.values = make([]float64, len(vals))
	copy(lc.values, vals)
}

type LinearDiagram struct {
	Title      string
	Width      int
	Height     int
	Grid       bool

	MinValue   float64
	MaxValue   float64
	Step       float64

	categories map[string]*LinearCategory
	labels     []string
}

func (d *LinearDiagram) NewCategory(name string) (cat *LinearCategory) {
	n := new(LinearCategory)

	n.values = make([]float64, 0)
	d.categories[name] = n

	return n
}



func (d *LinearDiagram) SetLabels(labels []string) {

	d.labels = make([]string, len(labels))
	copy(d.labels, labels)
}

func (d *LinearDiagram) calcMinMax() {

	for _, cat := range d.categories {
		for iVal := 0; iVal < len(cat.values); iVal++ {
			if d.MaxValue < cat.values[iVal] {
				d.MaxValue = cat.values[iVal]
			}
			if d.MinValue > cat.values[iVal] {
				d.MinValue = cat.values[iVal]
			}
		}
	}
}

func (d *LinearDiagram) build(w io.Writer) (err error) {

	d.calcMinMax()

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
	xStep := (d.Width - dsMarginLeft - dsMarginRight) / (lenLabels - 1)
	left := dsMarginLeft
	s.Group(fmt.Sprintf("text-anchor:middle;font-size:%d;fill:%s", dsLabelsFontSize, dsLabelsFontColor))
	for i := 0; i < lenLabels; i++ {
		s.Text(left, d.Height-dsMarginBottom+dsLabelsMargin, d.labels[i])
		left += xStep
	}
	s.Gend()

	// **************

	// TODO func validate()
	if d.Step <= 0 {
		err = errors.New(fmt.Sprintf("Error: Invalid VStep value '%d'", d.Step))
		return
	}

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
	var stepsCount int = int(valSegment/d.Step + 0.5) + 1
	var stepHeight int = graphHeight/(stepsCount-1)

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

		// Vertical grid
		left = dsMarginLeft + xStep
		for i := 1; i < lenLabels; i++ {

			s.Line(left, dsMarginTop, left, d.Height-dsMarginBottom)
			left += xStep
		}

		// Horizontal grid
		top = d.Height - dsMarginBottom - stepHeight
		for i := 1; i < stepsCount; i++ {
			s.Line(dsMarginLeft, top, d.Width-dsMarginRight, top)
			top -= stepHeight
		}

		s.Gend()
	}

	// Draw linear graphs and legend

	// Calculate height and start for legend
	lHeight := (dsMarginBottom - dsLabelsMargin) / (len(d.categories) + 1)
	lTop := d.Height - dsMarginBottom + dsLabelsMargin + lHeight/2

	for name, cat := range d.categories {

		s.Group(fmt.Sprintf("stroke-width:%d;stroke:%s", cat.LineWidth, cat.Color))

		x1 := dsMarginLeft
		//y1 := d.Height - dsMarginBottom - int((cat.values[0] - d.MinValue) * pxInVal)
		var multiplier float64 = float64(stepHeight)/d.Step

		var pointValue float64 = cat.values[0] - d.MinValue
		var stepsInPointValue int = int(pointValue/d.Step)
		var remain int = int((pointValue - float64(stepsInPointValue) * d.Step) * multiplier)

		y1 := d.Height - dsMarginBottom - int(pointValue/d.Step) * stepHeight - remain

		lenVals := len(cat.values) - 1
		if (lenLabels - 1) < lenVals {
			lenVals = lenLabels - 1
		}

		for iVal := 0; iVal < lenVals; iVal++ {

			x2 := dsMarginLeft+(iVal+1)*xStep

			pointValue = cat.values[iVal+1] - d.MinValue
			stepsInPointValue := int(pointValue/d.Step)
			remain  = int((pointValue - float64(stepsInPointValue) * d.Step) * multiplier)

			y2 := d.Height - dsMarginBottom - int(pointValue/d.Step) * stepHeight - remain

			s.Line(x1, y1, x2, y2)

			y1 = y2
			x1 = x2
		}
		s.Gend()

		// Draw legend
		// TODO draw legend in any side
		// TODO do not draw legend if it's do not fit?
		s.Rect(d.Width/2, lTop + lHeight/2 - dsLegendMarkSize/2, dsLegendMarkSize, dsLegendMarkSize,
			fmt.Sprintf("fill:%s", cat.Color))
		s.Text(d.Width/2 + dsLegendMarkSize + 5, lTop + lHeight/2, name,
			fmt.Sprintf("alignment-baseline:middle;font-size:%d;fill:%s", dsLegendFontSize, dsLabelsFontColor))
		lTop += lHeight

	}


	s.End()

	return
}