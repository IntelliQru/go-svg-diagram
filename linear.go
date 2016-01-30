package svgd
import (
	"io"
	"errors"
	"fmt"
	"github.com/ajstarks/svgo"
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
	Title  string
	Width  int
	Height int
	Grid bool
	VStep  int

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


func (d *LinearDiagram) build(w io.Writer) (err error) {

	var minValue, maxValue float64
	isFirst := true

	for name, cat := range d.categories {
		if len(cat.values) != len(d.labels) {
			err = errors.New(fmt.Sprintf("Error: Count of values for category '%s' does not match labels count",
				name))
			return
		}

		for iVal := 0; iVal < len(cat.values); iVal++ {
			if maxValue < cat.values[iVal] {
				maxValue = cat.values[iVal]
			}
			if isFirst {
				minValue = cat.values[iVal]
				isFirst = false
			} else if minValue > cat.values[iVal] {
				minValue = cat.values[iVal]
			}
		}
	}

	s := svg.New(w)
	s.Start(d.Width, d.Height)

	// Title
	s.Text(d.Width/2, dsMarginTop/2, d.Title,
		fmt.Sprintf("text-anchor:middle;font-size:%dpx;fill:%s", dsTitleFontSize, dsTitleFontColor))

	// Y axis
	s.Line(dsMarginLeft, d.Height-dsMarginBottom, dsMarginLeft, dsMarginTop,
		fmt.Sprintf("stroke-width:%d;stroke:%s;", dsAxisLineWidth, dsAxisLineColor))
	// X axis
	s.Line(dsMarginLeft, d.Height-dsMarginBottom, d.Width-dsMarginRight, d.Height-dsMarginBottom,
		fmt.Sprintf("stroke-width:%d;stroke:%s;", dsAxisLineWidth, dsAxisLineColor))

	// Write labels
	length := len(d.labels)
	xStep := (d.Width - dsMarginLeft - dsMarginRight) / (length - 1)
	left := dsMarginLeft

	s.Group(fmt.Sprintf("text-anchor:middle;font-size:%d;fill:%s", dsLabelsFontSize, dsLabelsFontColor))
	for i := 0; i < length; i++ {
		s.Text(left, d.Height-dsMarginBottom+dsLabelsMargin, d.labels[i])
		left += xStep
	}
	s.Gend()

	// Write Y values
	// TODO round Y start for step nearest value
	if d.VStep <= 0 {
		err = errors.New(fmt.Sprintf("Error: Invalid VStep value '%d'", d.VStep))
		return
	}

	yCount := float64(maxValue-minValue) * 1.1 / float64(d.VStep)
	vCount := int(yCount + 0.5)
	yStep := int(float64(d.Height-dsMarginTop-dsMarginBottom) / yCount)
	val := int(minValue)
	top := d.Height - dsMarginBottom

	s.Group(fmt.Sprintf("text-anchor:end;font-size:%d;fill:%s",
		dsLabelsFontSize, dsLabelsFontColor))
	for i := 0; i < vCount; i++ {
		s.Text(dsMarginLeft-dsValuesMargin, top, fmt.Sprintf("%d", val), "alignment-baseline:central")
		val += d.VStep
		top -= yStep
	}
	s.Gend()

	// Drawing grid
	if d.Grid {

		s.Group("stroke-width:1;stroke:lightgray")

		// Vertical grid
		left = dsMarginLeft + xStep
		for i := 1; i < length; i++ {

			s.Line(left, dsMarginTop, left, d.Height-dsMarginBottom)
			left += xStep
		}

		// Horizontal grid
		top = d.Height - dsMarginBottom - yStep
		for i := 1; i < vCount; i++ {
			s.Line(dsMarginLeft, top, d.Width-dsMarginRight, top)
			top -= yStep
		}

		s.Gend()
	}

	// Draw linear graphs

	lHeight := (dsMarginBottom - dsLabelsMargin - dsLabelsFontSize) / (len(d.categories) + 1)
	lTop := d.Height - dsMarginBottom + dsLabelsMargin + lHeight/2

	for name, cat := range d.categories {

		s.Group(fmt.Sprintf("stroke-width:%d;stroke:%s", dsLinearWidth, cat.Color))

		valLength := maxValue - minValue
		yLength := float64(yStep*(vCount-1))
		x1 := dsMarginLeft
		y1 := d.Height - dsMarginBottom - int((cat.values[0]-minValue)/valLength*yLength)

		length = len(cat.values)-1
		for iVal := 0; iVal < length; iVal++ {

			x2 := dsMarginLeft+(iVal+1)*xStep
			y2 := d.Height - dsMarginBottom - int((cat.values[iVal+1]-minValue)/valLength*yLength)

			s.Line(x1, y1, x2, y2)
			//s.Qbez(x1, y1, x2, y1, x2, y2)

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
			fmt.Sprintf("alignment-baseline:central;font-size:%d;fill:%s", dsLegendFontSize, dsLabelsFontColor))
		lTop += lHeight

	}


	s.End()

	return
}