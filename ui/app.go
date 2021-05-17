package ui

import (
	"errors"
	"sync/atomic"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/workers"
)

var App *Application

type AppState int

const (
	STATE_INIT AppState = iota
	STATE_FETCHING_MAILBOX
)

// transition info
type Tr struct {
	Msg workers.Message
    ImapCb func(res workers.Message) error
    DbCb func(res workers.Message) error
}

type Application struct {
	exit     atomic.Value // bool
	logger   *lib.Logger
	screen   tcell.Screen
	window   *Window
	// style    tcell.Style
	tcEvents chan tcell.Event
	cbId int
	done chan struct{}

	state AppState
	dbcallbacks map[int]func(m workers.Message) error
	imapcallbacks map[int]func(m workers.Message) error

	db *workers.Database
}

func InitApp(l *lib.Logger) error {
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
			logger: l,
			tcEvents: make(chan tcell.Event, 10),
			dbcallbacks: make(map[int]func(m workers.Message) error),
			imapcallbacks: make(map[int]func(m workers.Message) error),
			db: workers.NewDatabase(l),
			done: make(chan struct{}),
			screen: screen,
		}
		w := NewWindow()
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
	case tev := <-app.tcEvents:
		more = true
		switch tev.(type) {
		case *tcell.EventResize:
			app.screen.Sync()
			// XXX propagate resize to window?
			return true
		}
		app.window.TabHandleEvent(tev)
	default:
		if app.window.ShouldRedraw() {
			app.window.Redraw()
		}
	}
	return more
}

func (app *Application) PostDbMessage(req *Tr) {
	msg := req.Msg
	app.cbId++
	msg.SetId(app.cbId)
	if req.DbCb != nil {
		app.dbcallbacks[app.cbId] = req.DbCb
	}
	app.db.PostMessage(msg)
}


func (app *Application) Transition(req *Tr) {
	switch app.state {
	case STATE_INIT:
		switch req.Msg.(type) {
		case *workers.FetchMailbox:
			app.PostDbMessage(req)
			app.state = STATE_FETCHING_MAILBOX
		}
	}
}
// func (app *Application) HandleUiRequest(req *UiRequest) {
// 	switch app.state {
// 	case STATE_INIT:
// 		switch req := req.(type) {
// 		case *FetchMailboxes:
// 			app.PostDbMessage(req, func(res *Response) error {
// 				res, ok := res.(MailboxesResponse)
// 				if !ok {
// 					return fmt.Errorf("error fetching mailboxes")
// 				}
// 				l := widgets.NewList(flog)
// 				for _, m := range res.Mailboxes {
// 					l.AddLine(&Mailbox{title: m})
// 				}
// 				l.OnSelect = func(line widgets.IRune) {
// 					m, _ := line.(*Mailbox)
// 					app.PostRequest(&FetchThreads{Mailbox: m.title})
// 				}
// 			})
// 			app.state = STATE_FETCHING_MAILBOXES
// 		}
// 	case STATE_FETCHING_MAILBOXES:
// 	}
// }

func (app *Application) Run() {
	if err := app.initialize(); err != nil {
		panic(err)
	}
	go app.db.Run()

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
		// case req := <-app.uiRequest:
		// 	app.HandleUiRequest(req)
		// case appEv := <-app.uiEvents:
		// 	if appEv == widgets.QUIT_EVENT {
		// 		return
		// 	} else {
		// 		if appEv.cb != nil {
		// 			app.callbacks[appEv.getId()] = appEv.cb
		// 		}
		// 		app.db.PostMessage(appEv)
		// 	}
		case res := <-app.db.Responses():
			id := res.GetId()
			cb, ok := app.dbcallbacks[id]
			if !ok {
				app.logger.Warnf("cannot found callbacks with id %d", id)
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
