package main

import (
	"github.com/IntelliQru/go-svg-diagram"
	"log"
	"net/http"

)

func main() {
	http.Handle("/linear", http.HandlerFunc(linear))
	http.Handle("/bar", http.HandlerFunc(bar))
	http.Handle("/pie", http.HandlerFunc(pie))
	err := http.ListenAndServe(":2003", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func linear(w http.ResponseWriter, req *http.Request) {

	dg := svgd.NewDiagram()
	ld := dg.CreateLinear()

	ld.SetLabels([]string{"Январь", "Февраль", "Март", "Апрель", "Май", "Июнь", "Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь"})

	ld.Title = "Заголовок диаграммы";
	ld.Step = 50;
	ld.Width = 900;
	ld.Height = 600;
	ld.Grid = true;
	ld.MinValue = 100
	ld.MaxValue = 600

	cat := ld.NewCategory("Выручка 2014")
	cat.SetValues([]float64{-25, 460, 100})

	cat = ld.NewCategory("Выручка 2015")
	cat.SetValues([]float64{234, 23, 345, 76, 267})

	cat = ld.NewCategory("Выручка 2014")
	cat.SetValues([]float64{368, -10, 100, 451, 589, 99})

	cat = ld.NewCategory("Выручка 2015")
	cat.LineWidth = 2
	cat.SetValues([]float64{34, 765, 367, 796, 234, 235, 342, 23, 23, 345, 456, -300, 456, 34, 56, 345, 56, 98, 123, 345, 234, 234, 234, 345})

	err := dg.Build(w)

	/*
		var b bytes.Buffer
		wr := bufio.NewWriter(&b)
		dg.Build(wr)
		wr.Flush()

		fmt.Println(b.String())
	*/

	if err != nil {
		w.Write([]byte(err.Error()))
	}
}

func bar(w http.ResponseWriter, req *http.Request) {

	dg := svgd.NewDiagram()
	ld := dg.CreateBar()

	ld.SetLabels([]string{"2014", "2015", "2016"})

	ld.Title = "Заголовок диаграммы";
	ld.Step = 50;
	ld.Width = 900;
	ld.Height = 600;
	ld.Grid = true;
	ld.MinValue = -300
	ld.MaxValue = 1000

	cat := ld.NewCategory("Выручка 2014")

	cat.Color = "red"
	cat.SetValues([]float64{-25, 460, 100})

	cat = ld.NewCategory("Выручка 2015")
	cat.Color = "green"
	cat.SetValues([]float64{34, 765, 45})

	cat = ld.NewCategory("Выручка 2015")
	cat.SetValues([]float64{34, 765, 230})

	cat = ld.NewCategory("Выручка 2015")
	cat.SetValues([]float64{12, 560, 765})

	err := dg.Build(w)

	if err != nil {
		w.Write([]byte(err.Error()))
	}
}

func pie(w http.ResponseWriter, req *http.Request) {
	dg := svgd.NewDiagram()
	d := dg.CreatePie()
	d.Title = "Заголовок диаграммы";
	d.Width = 700;
	d.Height = 500;
	d.ShowValues = true
	d.ValuesShift = 180

	var shift uint = 4

	cat := d.NewCategory("Category 1")
	cat.Value = 100
	cat.Shift = shift

	cat = d.NewCategory("Category 2")
	cat.Value = 200
	cat.Shift = shift

	cat = d.NewCategory("Category 3")
	cat.Value = 300
	cat.Shift = shift

	cat = d.NewCategory("Category 4")
	cat.Value = 12
	cat.Shift = shift

	cat = d.NewCategory("Category 5")
	cat.Value = 24
	cat.Shift = shift

	cat = d.NewCategory("Category 6")
	cat.Value = 57
	cat.Shift = shift

	cat = d.NewCategory("Category 7")
	cat.Value = 99
	cat.Shift = shift

	err := dg.Build(w)

	if err != nil {
		w.Write([]byte(err.Error()))
	}
}
