package ui

import (
	"fmt"
	"sync/atomic"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"

	"github.com/stregouet/nuntius/config"
	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/models"
	"github.com/stregouet/nuntius/widgets"
	"github.com/stregouet/nuntius/workers"
)

const (
	STATE_SHOW_TAB lib.StateType      = "SHOW_TAB"
	TR_OPEN_TAB    lib.TransitionType = "OPEN_TAB"
	TR_CLOSE_TAB   lib.TransitionType = "CLOSE_TAB"
	TR_NEXT_TAB    lib.TransitionType = "NEXT_TAB"
	TR_PREV_TAB    lib.TransitionType = "PREV_TAB"

	STATE_WRITE_CMD  lib.StateType      = "WRITE_CMD"
	TR_START_WRITING lib.TransitionType = "START_WRITING"
	TR_CANCEL        lib.TransitionType = "CANCEL"
	TR_VALIDATE      lib.TransitionType = "VALIDATE"
)

type WindowMachineCtx struct {
	tabs        []widgets.Widget
	selectedTab int
}

func buildWindowMachine() *lib.Machine {
	return lib.NewMachine(
		&WindowMachineCtx{
			tabs:        make([]widgets.Widget, 0),
			selectedTab: 0,
		},
		STATE_SHOW_TAB,
		lib.States{
			STATE_SHOW_TAB: &lib.State{
				Transitions: lib.Transitions{
					TR_OPEN_TAB: &lib.Transition{
						Target: STATE_SHOW_TAB,
						Action: func(c interface{}, ev *lib.Event) {
							wmc := c.(*WindowMachineCtx)
							newwidget := ev.Payload.(widgets.Widget)
							wmc.tabs = append(wmc.tabs, newwidget)
							wmc.selectedTab = len(wmc.tabs) - 1
						},
					},
					TR_CLOSE_TAB:     &lib.Transition{Target: STATE_SHOW_TAB},
					TR_NEXT_TAB:      &lib.Transition{Target: STATE_SHOW_TAB},
					TR_PREV_TAB:      &lib.Transition{Target: STATE_SHOW_TAB},
					TR_START_WRITING: &lib.Transition{Target: STATE_WRITE_CMD},
				},
			},
			STATE_WRITE_CMD: &lib.State{
				Transitions: lib.Transitions{
					TR_CANCEL:   &lib.Transition{Target: STATE_SHOW_TAB},
					TR_VALIDATE: &lib.Transition{Target: STATE_SHOW_TAB},
				},
			},
		},
	)
}

type Window struct {
	screen tcell.Screen

	machine *lib.Machine
	ex      *Status

	triggerRedraw atomic.Value // bool
}

func NewWindow(cfg []*config.Account) *Window {
	w := &Window{
		machine: buildWindowMachine(),
		ex:      NewStatus("ici c'est pour les commandes"),
	}
	w.machine.OnTransition(func(s *lib.State, ev *lib.Event) {
		if ev.Transition == TR_OPEN_TAB {
			w.onOpenTab(s, ev)
		}
	})
	w.ex.AskingRedraw(func() {
		w.AskRedraw()
	})
	w.ResetRedraw()

	for _, c := range cfg {
		App.PostImapMessage(
			&workers.ConnectImap{},
			c.Name,
			func(response workers.Message) error {
				if r, ok := response.(*workers.Error); ok {
					w.ShowMessagef("cannot connect to imap server: %v", r.Error)
				}
				return nil
			},
		)
		accwidget := NewMailboxesView(c.Name, w.onSelectMailbox)
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
		w.machine.Send(&lib.Event{TR_OPEN_TAB, accwidget})
	}
	return w
}

func (w *Window) state() *WindowMachineCtx {
	return w.machine.Context.(*WindowMachineCtx)
}

func (w *Window) onSelectMailbox(acc string, mailbox *models.Mailbox) {
	mv := NewMailboxView(acc, nil)
	App.PostMessage(
		&workers.FetchMailbox{Mailbox: mailbox.Name},
		acc,
		func(response workers.Message) error {
			switch r := response.(type) {
			case *workers.Error:
				App.logger.Errorf("fetchmailbox res %v", response)
				w.ShowMessage(r.Error.Error())
			case *workers.FetchMailboxRes:
				mv.SetThreads(r.List)
			default:
				App.logger.Error("unknown response type")
			}
			return nil
		})
	w.machine.Send(&lib.Event{TR_OPEN_TAB, mv})
}

func (w *Window) onOpenTab(s *lib.State, ev *lib.Event) {
	widget := ev.Payload.(widgets.Widget)
	widget.AskingRedraw(func() {
		w.AskRedraw()
	})
	if w.screen != nil {
		w.screen.Clear()
		widget.SetViewPort(w.tabViewPort())
	}
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
	state := w.state()
	for _, t := range state.tabs {
		t.SetViewPort(w.tabViewPort())
	}
}
func (w *Window) Draw() {
	s := w.state()
	s.tabs[s.selectedTab].Draw()
	w.ex.Draw()
}
func (w *Window) TabHandleEvent(ev tcell.Event) {
	s := w.state()
	s.tabs[s.selectedTab].HandleEvent(ev)
}
func (w *Window) ExHandleEvent(ev tcell.Event) {
	w.ex.HandleEvent(ev)
}
