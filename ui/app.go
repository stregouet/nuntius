package ui

import (
	"log"
	"sync/atomic"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/stregouet/nuntius/widgets"
)

type Application struct {
	exit     atomic.Value // bool
	logger   *log.Logger
	screen   tcell.Screen
	window   *Window
	style    tcell.Style
	tcEvents chan tcell.Event
	uiEvents chan widgets.AppEvent
}

func NewApp(l *log.Logger) Application {
	app := Application{
		uiEvents: make(chan widgets.AppEvent, 10),
		logger:    l,
		tcEvents:  make(chan tcell.Event, 10),
	}
	app.exit.Store(false)
	return app
}

func (app *Application) SetStyle(style tcell.Style) {
	app.style = style
	if app.screen != nil {
		app.screen.SetStyle(style)
	}
}

func (app *Application) SetWindow(w *Window) {
	app.window = w
}

func (app *Application) initialize() error {
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	var err error
	if app.screen == nil {
		if app.screen, err = tcell.NewScreen(); err != nil {
			return err
		}
		app.screen.SetStyle(app.style)
	}
	return err
}

func (app *Application) exiting() bool {
	return app.exit.Load().(bool)
}

func (app *Application) Close() {
	app.screen.Fini()
}

func (app *Application) HandleUiEvent(ev widgets.AppEvent) {
	app.uiEvents <- ev
}

func (app *Application) tick() bool {
	more := false
	select {
	case tev := <-app.tcEvents:
		more = true
		switch tev.(type) {
		case *tcell.EventResize:
			app.screen.Sync()
			app.window.Resize()
			return true
		}
		app.window.HandleEvent(tev)
	default:
	}
	return more
}


func (app *Application) Run() {
	if err := app.initialize(); err != nil {
		panic(err)
	}
	app.screen.Init()
	app.screen.Clear()
	app.window.SetScreen(app.screen)
	go func() {
		for !app.exiting() {
			app.tcEvents <- app.screen.PollEvent()
		}
	}()
	defer app.Close()
	app.window.Draw()
	app.screen.Show()
	for {
		select {
		case appEv := <-app.uiEvents:
			if appEv == widgets.QUIT_EVENT {
				return
			}
		default:
			for app.tick() {

			}
			time.Sleep(16 * time.Millisecond)
		}
	}
}
