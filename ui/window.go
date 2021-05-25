package ui

import (
	"fmt"
	"sync/atomic"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"

	"github.com/stregouet/nuntius/config"
	"github.com/stregouet/nuntius/widgets"
	"github.com/stregouet/nuntius/workers"
)

type Window struct {
	screen tcell.Screen

	selectedTab int
	tabs        []widgets.Widget
	ex          *Status

	triggerRedraw atomic.Value // bool
}

func NewWindow(cfg []*config.Account) *Window {
	w := &Window{
		tabs: make([]widgets.Widget, 0),
		ex:   NewStatus("ici c'est pour les commandes"),
	}
	w.ex.AskingRedraw(func() {
		w.AskRedraw()
	})
	w.ResetRedraw()


	for _, c := range cfg {
		App.PostImapMessage(
			&workers.ConnectImap{},
			c.Name,
			func(response workers.Message) error {
				if  r, ok := response.(*workers.Error); ok {
					w.ShowMessagef("cannot connect to imap server: %v", r.Error)
				}
				return nil
			},
		)
		accwidget := NewMailboxesView(c.Name)
		App.PostMessage(
			&workers.FetchMailboxes{},
			c.Name,
			func(response workers.Message) error {
				switch r := response.(type) {
				case *workers.Error:
					App.logger.Errorf("fetchmailboxes res %v", response)
					w.ShowMessage(r.Error.Error())
				case *workers.FetchMailboxesRes:
					App.logger.Debugf("fetchmailboxes res %v", response)
					accwidget.SetMailboxes(r.Mailboxes) // XXX should call askredraw
				default:
					App.logger.Error("unknown response type")
				}
				return nil
			})
		w.AddTab(accwidget)
	}
	return w
}

func (w *Window) tabViewPort() *views.ViewPort {
	_, h := w.screen.Size()
	return views.NewViewPort(w.screen, 0, 0, -1, h-1)
}
func (w *Window) exViewPort() *views.ViewPort {
	_, h := w.screen.Size()
	return views.NewViewPort(w.screen, 0, h-1, -1, -1)
}

func (w *Window) ShowMessage(msg string) {
	w.ex.ShowMessage(msg)
}
func (w *Window) ShowMessagef(msg string, args ...interface{}) {
	m := fmt.Sprintf(msg, args...)
	w.ex.ShowMessage(m)
}

func (w *Window) AddTab(widget widgets.Widget) {
	w.tabs = append(w.tabs, widget)
	widget.AskingRedraw(func() {
		w.AskRedraw()
	})
	w.selectedTab = len(w.tabs) - 1
	if w.screen != nil {
		widget.SetViewPort(w.tabViewPort())
	}
}

func (w *Window) ShouldRedraw() bool {
	return w.triggerRedraw.Load().(bool)
}
func (w *Window) AskRedraw() {
	w.triggerRedraw.Store(true)
}
func (w *Window) ResetRedraw() {
	w.triggerRedraw.Store(false)
}
func (w *Window) Redraw() {
	w.ResetRedraw()
	w.Draw()
	w.screen.Show()
}

func (w *Window) Size() (int, int) { return w.screen.Size() }
func (w *Window) SetScreen(s tcell.Screen) {
	w.screen = s
	w.ex.SetViewPort(w.exViewPort())
	for _, t := range w.tabs {
		t.SetViewPort(w.tabViewPort())
	}
}
func (w *Window) Draw() {
	w.tabs[w.selectedTab].Draw()
	w.ex.Draw()
}
func (w *Window) TabHandleEvent(ev tcell.Event) {
	w.tabs[w.selectedTab].HandleEvent(ev)
}
func (w *Window) ExHandleEvent(ev tcell.Event) {
	w.ex.HandleEvent(ev)
}
