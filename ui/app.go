package ui

import (
	"log"
	"sync/atomic"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
)

type Application struct {
	AppEvents chan (AppEvent)

	exit     atomic.Value // bool
	logger   *log.Logger
	screen   tcell.Screen
	widget   views.Widget
	style    tcell.Style
	tcEvents chan tcell.Event
}

func NewApp(l *log.Logger) Application {
	app := Application{
		AppEvents: make(chan AppEvent, 10),
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

func (app *Application) SetWidget(widget views.Widget) {
	app.widget = widget
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

func (app *Application) tick() bool {
	more := false
	select {
	case tev := <-app.tcEvents:
		app.logger.Print("dans tick recu un tcell event")
		more = true
		// switch ev:=tev.(type) {
		switch tev.(type) {
		case *tcell.EventResize:
			app.screen.Sync()
			app.widget.Resize()
			return true
		}
		app.logger.Print("event pour app.widget")
		app.widget.HandleEvent(tev)
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
	app.widget.SetView(app.screen)
	go func() {
		for !app.exiting() {
			app.tcEvents <- app.screen.PollEvent()
		}
	}()
	defer app.Close()
	for {
		app.widget.Draw()
		app.screen.Show()
		// w, h := app.screen.Size()
		// app.logger.Printf("screen size w:%d, h:%d", w, h)

		select {
		case appEv := <-app.AppEvents:
			app.logger.Printf("\tfound one app event %d", appEv)
			switch appEv {
			case QUIT_EVENT:
				return
			}
		default:
			for app.tick() {

			}
			time.Sleep(16 * time.Millisecond)
		}
	}
}
