package svgd

import (
	"errors"
	"fmt"
	"github.com/ajstarks/svgo"
	"io"
)

type DiagramType int

const (
	DT_LINE DiagramType = iota
	DT_PIE
	DT_BAR
)

const (
	dsMarginLeft   = 50
	dsMarginRight  = 50
	dsMarginTop    = 50
	dsMarginBottom = 150

	dsAxisLineWidth = 2
	dsAxisLineColor = "darkgray"

	dsLabelsMargin = 20
	dsValuesMargin = 10

	dsTitleFontSize   = 20
	dsLabelsFontSize  = 12
	dsTitleFontColor  = "#3C3C3C"
	dsLabelsFontColor = "#3C3C3C"

	dsLinearWidth = 3
)

type DiagramSettings struct {
	Type   DiagramType
	Width  int
	Height int
	Title  string
	Grid   bool
	VStep  int
}

type DiagramCategory struct {
	Name   string
	Color  string
	Values []float64
}

type Diagram struct {
	dType  DiagramType
	title  string
	width  int
	height int

	isGrid bool
	vStep  int

	categories []DiagramCategory
	labels     []string
}

func NewDiagram(settings DiagramSettings) *Diagram {

	dg := Diagram{
		title:      settings.Title,
		dType:      settings.Type,
		width:      settings.Width,
		height:     settings.Height,
		categories: make([]DiagramCategory, 0),
		isGrid:     settings.Grid,
		vStep:      settings.VStep,
	}

	return &dg
}

func (d *Diagram) AddCategory(cat DiagramCategory) {

	vals := make([]float64, len(cat.Values))
	copy(vals, cat.Values)

	newCat := DiagramCategory{
		Name:   cat.Name,
		Color:  cat.Color,
		Values: vals,
	}

	d.categories = append(d.categories, newCat)

}

func (d *Diagram) SetLabels(labels []string) {

	d.labels = make([]string, len(labels))
	copy(d.labels, labels)
}

func (d *Diagram) Build(w io.Writer) (err error) {

	var minValue, maxValue float64
	isFirst := true

	for iCat := 0; iCat < len(d.categories); iCat++ {
		if len(d.categories[iCat].Values) != len(d.labels) {
			err = errors.New(fmt.Sprintf("Error: Count of values for category '%s' does not match labels count",
				d.categories[iCat].Name))
			return
		}

		for iVal := 0; iVal < len(d.categories[iCat].Values); iVal++ {
			if maxValue < d.categories[iCat].Values[iVal] {
				maxValue = d.categories[iCat].Values[iVal]
			}
			if isFirst {
				minValue = d.categories[iCat].Values[iVal]
				isFirst = false
			} else if minValue > d.categories[iCat].Values[iVal] {
				minValue = d.categories[iCat].Values[iVal]
			}
		}
	}

	s := svg.New(w)
	s.Start(d.width, d.height)

	// Title
	s.Text(d.width/2, dsMarginTop/2, d.title,
		fmt.Sprintf("text-anchor:middle;font-size:%d;fill:%s", dsTitleFontSize, dsTitleFontColor))

	// Y axis
	s.Line(dsMarginLeft, d.height-dsMarginBottom, dsMarginLeft, dsMarginTop,
		fmt.Sprintf("stroke-width:%d;stroke:%s;", dsAxisLineWidth, dsAxisLineColor))
	// X axis
	s.Line(dsMarginLeft, d.height-dsMarginBottom, d.width-dsMarginRight, d.height-dsMarginBottom,
		fmt.Sprintf("stroke-width:%d;stroke:%s;", dsAxisLineWidth, dsAxisLineColor))

	// Write labels
	length := len(d.labels)
	xStep := (d.width - dsMarginLeft - dsMarginRight) / (length - 1)
	left := dsMarginLeft

	s.Group(fmt.Sprintf("text-anchor:middle;font-size:%d;fill:%s", dsLabelsFontSize, dsLabelsFontColor))
	for i := 0; i < length; i++ {
		s.Text(left, d.height-dsMarginBottom+dsLabelsMargin, d.labels[i])
		left += xStep
	}
	s.Gend()

	// Write Y values
	// TODO round Y start for step nearest value
	if d.vStep <= 0 {
		err = errors.New(fmt.Sprintf("Error: Invalid VStep value '%d'", d.vStep))
		return
	}

	yCount := float64(maxValue-minValue) * 1.1 / float64(d.vStep)
	vCount := int(yCount + 0.5)
	yStep := int(float64(d.height-dsMarginTop-dsMarginBottom) / yCount)
	val := int(minValue)
	top := d.height - dsMarginBottom

	s.Group(fmt.Sprintf("text-anchor:end;font-size:%d;fill:%s", dsLabelsFontSize, dsLabelsFontColor))
	for i := 0; i < vCount; i++ {
		s.Text(dsMarginLeft-dsValuesMargin, top, fmt.Sprintf("%d", val))
		val += d.vStep
		top -= yStep
	}
	s.Gend()

	// Drawing grid
	if d.isGrid {

		s.Group("stroke-width:1;stroke:lightgray")

		// Vertical grid
		left = dsMarginLeft + xStep
		for i := 1; i < length; i++ {

			s.Line(left, dsMarginTop, left, d.height-dsMarginBottom)
			left += xStep
		}

		// Horizontal grid
		top = d.height - dsMarginBottom - yStep
		for i := 1; i < vCount; i++ {
			s.Line(dsMarginLeft, top, d.width-dsMarginRight, top)
			top -= yStep
		}

		s.Gend()
	}

	// Draw linear graphs

	for iCat := 0; iCat < len(d.categories); iCat++ {

		s.Group(fmt.Sprintf("stroke-width:%d;stroke:%s", dsLinearWidth, d.categories[iCat].Color))

		valLength := maxValue - minValue
		yLength := float64(yStep*(vCount-1))
		x1 := dsMarginLeft
		y1 := d.height - dsMarginBottom - int((d.categories[iCat].Values[0]-minValue)/valLength*yLength)

		length = len(d.categories[iCat].Values)-1
		for iVal := 0; iVal < length; iVal++ {

			x2 := dsMarginLeft+(iVal+1)*xStep
			y2 := d.height - dsMarginBottom - int((d.categories[iCat].Values[iVal+1]-d.categories[iCat].Values[0])/valLength*yLength)

			s.Line(x1, y1, x2, y2)
			fmt.Println(valLength, yLength, x1, y1, x2, y2)

			y1 = y2
			x1 = x2
		}
		s.Gend()
	}

	s.End()

	return
}
