package svgd

import (
	"errors"
	"io"
)

const (
	dsMarginLeft   = 80
	dsMarginRight  = 100
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

	dsLegendMarkSize  = 15
	dsLegendFontSize  = 10
	dsLegendMargin = 10
	dsLegendFontColor = "#3C3C3C"

	dsBarMargin       = 5
	dsPieMargin		  = 25
)

type diagramInterface interface {
	build(w io.Writer) error
}

type Diagram struct {
	diagram diagramInterface
}

func NewDiagram() *Diagram {

	dg := Diagram{}

	return &dg
}

func (d *Diagram) CreateLinear() (dg *LinearDiagram) {
	newLD := new(LinearDiagram)
	newLD.categories = make([]*LinearCategory, 0)

	d.diagram = newLD

	return newLD
}

func (d *Diagram) CreateBar() (dg *BarDiagram) {
	newLD := new(BarDiagram)
	newLD.categories = make([]*BarCategory, 0)

	d.diagram = newLD

	return newLD
}

func (d *Diagram) CreatePie() (dg *PieDiagram) {
	newLD := new(PieDiagram)
	newLD.categories = make([]*PieCategory, 0)

	d.diagram = newLD

	return newLD
}

func (d *Diagram) Build(w io.Writer) (err error) {

	if d.diagram == nil {
		err = errors.New("Error: No diagram created")
		return
	}
	err = d.diagram.build(w)

	return
}
