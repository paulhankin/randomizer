package main

import (
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"

	"gioui.org/app"
	"gioui.org/f32"
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

type StateOne struct {
	bg, fg color.NRGBA
	msg    string
}

type State struct {
	was, now  StateOne
	clickTime time.Time
	tloc      f32.Point
}

func rgb(r, g, b int) color.NRGBA {
	return color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	go func() {
		SX, SY := unit.Px(300), unit.Px(200)
		w := app.NewWindow(app.MaxSize(SX, SY), app.MinSize(SX, SY), app.Size(SX, SY), app.Title("Paul's randomizer"))
		state := &State{
			now:       StateOne{bg: rgb(0, 0, 0), fg: rgb(255, 255, 255), msg: "click"},
			clickTime: time.Now().Add(-100 * time.Second),
		}
		err := state.run(w)
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

func c1Lerp(c, d uint8, t float64) uint8 {
	return uint8(float64(c)*(1-t) + float64(d)*t)
}

func colorLerp(c, d color.NRGBA, t float64) color.NRGBA {
	r := c1Lerp(c.R, d.R, t)
	g := c1Lerp(c.G, d.G, t)
	b := c1Lerp(c.B, d.B, t)
	a := c1Lerp(c.A, d.A, t)
	return color.NRGBA{r, g, b, a}
}

const transitionTime = 500 * time.Millisecond

func (s *State) run(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())
	var ops op.Ops
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			released := false
			var ploc f32.Point
			for _, e := range gtx.Events(clickHandler) {
				if pe, ok := e.(pointer.Event); ok {
					if time.Now().Sub(s.clickTime) >= transitionTime {
						released = true
						ploc = pe.Position
					}
				}
			}
			if released {
				s.was = s.now
				rnd := rand.Intn(100) + 1
				col0 := rgb(255, 0, 0)
				col1 := rgb(200, 200, 0)
				col2 := rgb(0, 255, 0)
				var bg color.NRGBA
				if rnd <= 50 {
					bg = colorLerp(col0, col1, float64(rnd)/50)
				} else {
					bg = colorLerp(col1, col2, float64(rnd-50)/50)
				}
				fg := rgb(255-int(bg.R), 255-int(bg.G), 255-int(bg.B))
				fg = colorLerp(rgb(0, 0, 0), fg, 0.5)
				s.now = StateOne{
					bg:  bg,
					fg:  fg,
					msg: strconv.Itoa(rnd),
				}
				s.clickTime = time.Now()
				s.tloc = ploc
			}

			t := math.Min(1.0, float64(time.Now().Sub(s.clickTime))/float64(transitionTime))
			pointer.InputOp{Tag: clickHandler, Types: pointer.Release}.Add(gtx.Ops)
			if t < 1 {
				op.InvalidateOp{}.Add(gtx.Ops)
			}
			for i, st := range []StateOne{s.was, s.now, s.now} {
				if t == 1 && i < 2 {
					continue
				}
				var stack clip.Stack
				bgDarken := 1.0
				if i > 0 && t < 1 {
					mx := float64(gtx.Constraints.Max.X)
					my := float64(gtx.Constraints.Max.Y)
					D := math.Max(mx, my)
					tt := math.Pow(t, 5)
					radius := float32(D * tt)
					if i == 1 {
						radius *= 1.5
						bgDarken = 0.8
					}
					tx := s.tloc.X
					ty := s.tloc.Y
					emin := f32.Point{X: tx - radius, Y: ty - radius}
					emax := f32.Point{X: tx + radius, Y: ty + radius}
					stack = clip.Ellipse{Min: emin, Max: emax}.Push(gtx.Ops)
				}
				paint.ColorOp{Color: colorLerp(rgb(0, 0, 0), st.bg, bgDarken)}.Add(gtx.Ops)
				paint.PaintOp{}.Add(gtx.Ops)
				layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Flexed(1.0, func(gtx layout.Context) layout.Dimensions {
						return layout.Dimensions{Size: gtx.Constraints.Max}
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						title := material.H1(th, st.msg)
						title.Color = st.fg
						title.Alignment = text.Middle
						return title.Layout(gtx)
					}),
					layout.Flexed(1.0, func(gtx layout.Context) layout.Dimensions {
						return layout.Dimensions{Size: gtx.Constraints.Max}
					}),
				)
				if i > 0 && t < 1 {
					stack.Pop()
				}
			}
			e.Frame(gtx.Ops)
		}
	}
}
