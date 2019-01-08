package gomobileapp

import (
	"math/rand"
	"time"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/exp/sprite"
	"golang.org/x/mobile/exp/sprite/clock"
	"golang.org/x/mobile/exp/sprite/glsprite"
	"golang.org/x/mobile/gl"
)

// GoApp is go app handler
type GoApp struct {
	StartTime time.Time
	Images    *glutil.Images
	Eng       sprite.Engine
	Sz        size.Event

	onStart  func(gl.Context, sprite.Engine)
	onStop   func()
	onUpdate func(clock.Time) *sprite.Node

	onPaint func(gl.Context, size.Event)
	onTouch func(touch.Event)
	onKey   func(key.Event)
}

// StartGoApp is 4 starting go apps
func StartGoApp(ga GoApp, seed int64) {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	rand.Seed(seed)

	ga.StartTime = time.Now()

	app.Main(func(a app.App) {
		var glctx gl.Context
		var sz size.Event

		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			// 볼 수 있는 상태면, 그림 그린다
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					glctx, _ = e.DrawContext.(gl.Context)

					ga.Images = glutil.NewImages(glctx)
					ga.Eng = glsprite.Engine(ga.Images)
					ga.onStart(glctx, ga.Eng)

					a.Send(paint.Event{})
				case lifecycle.CrossOff:
					ga.onStop()
					ga.Eng.Release()
					ga.Images.Release()
					glctx = nil
				}
			case size.Event:
				sz = e
				ga.Sz = e
			case paint.Event:
				if glctx == nil || e.External {
					continue
				}
				glctx.ClearColor(1, 1, 1, 1)
				glctx.Clear(gl.COLOR_BUFFER_BIT)
				now := clock.Time(time.Since(ga.StartTime) * 60 / time.Second)
				scene := ga.onUpdate(now)
				ga.Eng.Render(scene, now, sz)
				ga.onPaint(glctx, sz)
				a.Publish()
				a.Send(paint.Event{}) // keep animating
			case touch.Event:
				ga.onTouch(e)
			case key.Event:
				ga.onKey(e)
			}
		}
	})
}
