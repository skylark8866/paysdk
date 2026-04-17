package main

import (
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

func main() {
	cfg := &Config{
		AppID:     getEnv("XGDN_APP_ID", ""),
		AppSecret: getEnv("XGDN_APP_SECRET", ""),
		// BaseURL:   getEnv("XGDN_BASE_URL", "http://localhost:8093"),
		BaseURL: getEnv("XGDN_BASE_URL", "https://pay.xgdn.net"),
	}

	state := NewState(cfg)
	ui := NewUI()

	go func() {
		w := new(app.Window)
		w.Option(app.Title("XGDN Pay 测试工具"))
		w.Option(app.Size(unit.Dp(480), unit.Dp(640)))
		w.Option(app.MinSize(unit.Dp(400), unit.Dp(500)))

		if err := loop(w, ui, state); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()

	app.Main()
}

func loop(w *app.Window, ui *UI, state *AppState) error {
	var ops op.Ops

	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			state.StopPolling()
			return e.Err

		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			ui.HandleEvents(gtx, state)

			background := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
			paint.Fill(gtx.Ops, background)

			ui.Layout(gtx, state)

			e.Frame(gtx.Ops)

		case key.Event:
			if e.Name == key.NameEscape {
				state.StopPolling()
				w.Perform(system.ActionClose)
			}
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
