package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math/rand"
	"os"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

func main() {
	go func() {
		w := app.NewWindow(app.MaxSize(unit.Px(300), unit.Px(200)), app.Title("Paul's randomizer"))
		err := run(w)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func ColorBox(gtx layout.Context, size image.Point, color color.NRGBA) layout.Dimensions {
	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: color}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	return layout.Dimensions{Size: size}
}

var clickHandler = new(int)

func run(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())
	var ops op.Ops
	rnd := 100
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			released := false
			for range gtx.Events(clickHandler) {
				released = true
			}
			if released {
				rnd = rand.Intn(100) + 1
			}
			pointer.InputOp{Tag: clickHandler, Types: pointer.Release}.Add(gtx.Ops)
			bg, fg := color.NRGBA{0, 0, 0, 255}, color.NRGBA{255, 255, 255, 255}
			if rnd <= 50 {
				bg, fg = fg, bg
			}
			paint.ColorOp{Color: bg}.Add(gtx.Ops)
			paint.PaintOp{}.Add(gtx.Ops)
			layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Flexed(1.0, func(gtx layout.Context) layout.Dimensions {
					return layout.Dimensions{Size: gtx.Constraints.Max}
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					title := material.H1(th, fmt.Sprintf("%d", rnd))
					title.Color = fg
					title.Alignment = text.Middle
					return title.Layout(gtx)
				}),
				layout.Flexed(1.0, func(gtx layout.Context) layout.Dimensions {
					return layout.Dimensions{Size: gtx.Constraints.Max}
				}),
			)
			e.Frame(gtx.Ops)
		}
	}
}
