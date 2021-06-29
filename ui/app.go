package ui

import (
	"errors"
	"sync/atomic"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/stregouet/nuntius/config"
	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/workers"
	"github.com/stregouet/nuntius/workers/imap"
)

var App *Application

type PostCallback func(res workers.Message) error

type Application struct {
	exit   atomic.Value // bool
	logger *lib.Logger
	screen tcell.Screen
	window *Window
	// style    tcell.Style
	transitions chan *lib.Event
	tcEvents chan tcell.Event
	cbId     int
	done     chan struct{}

	dbcallbacks   map[int]PostCallback
	imapcallbacks map[int]PostCallback

	db   *workers.Database
	imap *imap.ImapWorker
}

func InitApp(l *lib.Logger, cfg *config.Config) error {
	if App == nil {
		tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
		screen, err := tcell.NewScreen()
		if err != nil {
			return err
		}
		screen.SetStyle(tcell.StyleDefault.
			Foreground(tcell.ColorWhite).
			Background(tcell.ColorBlack))
		App = &Application{
			logger:        l,
			transitions:   make(chan *lib.Event, 10),
			tcEvents:      make(chan tcell.Event, 10),
			dbcallbacks:   make(map[int]PostCallback),
			imapcallbacks: make(map[int]PostCallback),
			db:            workers.NewDatabase(l),
			imap:          imap.NewImapWorker(l, cfg.Accounts),
			done:          make(chan struct{}),
			screen:        screen,
		}
		w := NewWindow(cfg.Accounts, cfg.Keybindings)
		w.SetScreen(screen)
		App.window = w
		App.exit.Store(false)
		return nil
	}
	return errors.New("InitApp should be called only once")
}

// func (app *Application) SetStyle(style tcell.Style) {
// 	app.style = style
// 	if app.screen != nil {
// 		app.screen.SetStyle(style)
// 	}
// }

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
		app.screen.SetStyle(tcell.StyleDefault.
			Foreground(tcell.ColorWhite).
			Background(tcell.ColorBlack))
	}
	return err
}

func (app *Application) exiting() bool {
	return app.exit.Load().(bool)
}

func (app *Application) Close() {
	app.screen.Fini()
}

func (app *Application) Stop() {
	close(app.done)
}

func (app *Application) tick() bool {
	more := false
	select {
	case ev := <-app.transitions:
		more = true
		app.window.HandleTransitions(ev)
	case tev := <-app.tcEvents:
		more = true
		switch tev.(type) {
		case *tcell.EventResize:
			app.screen.Sync()
			// XXX propagate resize to window?
			return true
		}
		app.window.HandleEvent(tev)
	default:
		if app.window.ShouldRedraw() {
			app.window.Redraw()
		}
	}
	return more
}

func (app *Application) PostDbMessage(msg workers.Message, accountname string, f PostCallback) {
	app.cbId++
	msg.SetId(app.cbId)
	msg.SetAccName(accountname)
	if f != nil {
		app.dbcallbacks[app.cbId] = f
	}
	app.db.PostMessage(msg)
}
func (app *Application) PostImapMessage(msg workers.Message, accountname string, f PostCallback) {
	app.cbId++
	msg.SetId(app.cbId)
	msg.SetAccName(accountname)
	if f != nil {
		app.imapcallbacks[app.cbId] = f
	}
	app.imap.PostMessage(msg)
}

func (app *Application) PostMessage(m workers.ClonableMessage, accountname string, f PostCallback) {
	app.PostDbMessage(m.Clone(), accountname, f)
	app.PostImapMessage(m, accountname, func(res workers.Message) error {
		if res, ok := res.(*workers.MsgToDb); ok {
			wrp := res.Wrapped
			app.PostDbMessage(wrp, accountname, f)
		}
		return nil
	})
}

func (app *Application) Run() {
	if err := app.initialize(); err != nil {
		panic(err)
	}
	go app.db.Run()
	go app.imap.Run()

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
		case <-app.done:
			return
		case res := <-app.imap.Responses():
			id := res.GetId()
			cb, ok := app.imapcallbacks[id]
			if !ok {
				app.logger.Warnf("cannot found imap callbacks with id %d", id)
			} else {
				cb(res)
				delete(app.imapcallbacks, id)
			}
		case res := <-app.db.Responses():
			id := res.GetId()
			cb, ok := app.dbcallbacks[id]
			if !ok {
				app.logger.Warnf("cannot found db callbacks with id %d", id)
			} else {
				cb(res)
				delete(app.dbcallbacks, id)
			}
		default:
			for app.tick() {

			}
			time.Sleep(16 * time.Millisecond)
		}
	}
}
