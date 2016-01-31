package svgd

import (
	"errors"
	"fmt"
	"github.com/ajstarks/svgo"
	"io"
	"math"
	"math/rand"
)

type LinearCategory struct {
	Name	  string
	Color     string
	LineWidth int
	values    []float64
}

func (lc *LinearCategory) SetValues(vals []float64) {

	lc.values = append(lc.values, vals...)
}

type LinearDiagram struct {
	Title  string
	Width  uint
	Height uint
	Grid   bool

	MinValue float64
	MaxValue float64
	Step     float64

	categories []*LinearCategory
	labels     []string
}

func (d *LinearDiagram) NewCategory(name string) (cat *LinearCategory) {
	n := new(LinearCategory)

	n.values = make([]float64, 0)
	n.Name = name
	d.categories = append(d.categories, n)

	return n
}

func (d *LinearDiagram) SetLabels(labels []string) {

	d.labels = make([]string, len(labels))
	copy(d.labels, labels)
}

func (d *LinearDiagram) validate() (err error) {

	if d.Step <= 0 {
		err = errors.New("Error: Step must be greater than zero")
	}

	if len(d.categories) == 0 {
		err = errors.New("Error: Nothing to build, categories are empty")
	}

	// Calculate Min and Max values
	for _, cat := range d.categories {
		for iVal := 0; iVal < len(cat.values); iVal++ {
			if d.MaxValue < cat.values[iVal] {
				d.MaxValue = cat.values[iVal]
			}
			if d.MinValue > cat.values[iVal] {
				d.MinValue = cat.values[iVal]
			}
		}
		if cat.LineWidth == 0 {
			cat.LineWidth = 1
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

func (d *LinearDiagram) build(w io.Writer) (err error) {

	if err = d.validate(); err != nil {
		return
	}

	s := svg.New(w)
	s.Start(int(d.Width), int(d.Height))

	// Title
	s.Text(int(d.Width)/2, dsMarginTop/2, d.Title,
		fmt.Sprintf("text-anchor:middle;alignment-baseline:central;font-size:%d;fill:%s",
			dsTitleFontSize, dsTitleFontColor))

	// Draw X and Y axis
	s.Line(dsMarginLeft, int(d.Height)-dsMarginBottom, int(d.Width)-dsMarginRight, int(d.Height)-dsMarginBottom,
		fmt.Sprintf("stroke-width:%d;stroke:%s;", dsAxisLineWidth, dsAxisLineColor))
	s.Line(dsMarginLeft, int(d.Height)-dsMarginBottom, dsMarginLeft, dsMarginTop,
		fmt.Sprintf("stroke-width:%d;stroke:%s;", dsAxisLineWidth, dsAxisLineColor))

	// Write labels
	lenLabels := len(d.labels)
	xStep := (int(d.Width) - dsMarginLeft - dsMarginRight) / (lenLabels - 1)
	left := dsMarginLeft
	s.Group(fmt.Sprintf("text-anchor:middle;font-size:%d;fill:%s", dsLabelsFontSize, dsLabelsFontColor))
	for i := 0; i < lenLabels; i++ {
		s.Text(left, int(d.Height)-dsMarginBottom+dsLabelsMargin, d.labels[i])
		left += xStep
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
	var graphHeight int = int(d.Height) - dsMarginBottom - dsMarginTop
	var valSegment float64 = d.MaxValue - d.MinValue
	var stepsCount int = int(valSegment/d.Step+0.5) + 1
	var stepHeight int = graphHeight / (stepsCount - 1)

	// Write Y values
	textValue := d.MinValue
	top := int(d.Height) - dsMarginBottom

	s.Group(fmt.Sprintf("text-anchor:end;font-size:%d;fill:%s",
		dsLabelsFontSize, dsLabelsFontColor))
	for i := 0; i < stepsCount; i++ {
		s.Text(dsMarginLeft-dsValuesMargin, top, fmt.Sprintf("%.2f", textValue))
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

			s.Line(left, dsMarginTop, left, int(d.Height)-dsMarginBottom)
			left += xStep
		}

		// Horizontal grid
		top = int(d.Height) - dsMarginBottom - stepHeight
		for i := 1; i < stepsCount; i++ {
			s.Line(dsMarginLeft, top, int(d.Width)-dsMarginRight, top)
			top -= stepHeight
		}

		s.Gend()
	}

	// Draw linear graphs and legend

	// Calculate height and start for legend
	lHeight := (dsMarginBottom - dsLabelsMargin) / (len(d.categories) + 1)
	lTop := int(d.Height) - dsMarginBottom + dsLabelsMargin + lHeight/2

	for _, cat := range d.categories {

		s.Group(fmt.Sprintf("stroke-width:%d;stroke:%s", cat.LineWidth, cat.Color))

		x1 := dsMarginLeft
		//y1 := int(d.Height) - dsMarginBottom - int((cat.values[0] - d.MinValue) * pxInVal)
		var multiplier float64 = float64(stepHeight) / d.Step

		var pointValue float64 = cat.values[0] - d.MinValue
		var stepsInPointValue int = int(pointValue / d.Step)
		var remain int = int((pointValue - float64(stepsInPointValue)*d.Step) * multiplier)

		y1 := int(d.Height) - dsMarginBottom - int(pointValue/d.Step)*stepHeight - remain

		lenVals := len(cat.values)
		if lenLabels < lenVals {
			lenVals = lenLabels
		}

		for iVal := 0; iVal < (lenVals - 1); iVal++ {

			x2 := dsMarginLeft + (iVal+1)*xStep

			pointValue = cat.values[iVal+1] - d.MinValue
			stepsInPointValue := int(pointValue / d.Step)
			remain = int((pointValue - float64(stepsInPointValue)*d.Step) * multiplier)

			y2 := int(d.Height) - dsMarginBottom - int(pointValue/d.Step)*stepHeight - remain

			s.Line(x1, y1, x2, y2)

			y1 = y2
			x1 = x2
		}
		s.Gend()

		// Draw legend
		// TODO draw legend in any side
		// TODO do not draw legend if it's do not fit?
		s.Rect(int(d.Width)/2, lTop+lHeight/2-dsLegendMarkSize/2, dsLegendMarkSize, dsLegendMarkSize,
			fmt.Sprintf("fill:%s", cat.Color))

		s.Text(int(d.Width)/2+dsLegendMarkSize+5, lTop+lHeight/2+dsLegendFontSize/2, cat.Name,
			fmt.Sprintf("font-size:%d;fill:%s", dsLegendFontSize, dsLabelsFontColor))
		lTop += lHeight

	}

	s.End()

	return
}
