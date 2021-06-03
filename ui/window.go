package ui

import (
	"fmt"
	"sync/atomic"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"

	"github.com/stregouet/nuntius/config"
	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/models"
	"github.com/stregouet/nuntius/workers"
)

const (
	STATE_SHOW_TAB lib.StateType      = "SHOW_TAB"
	TR_OPEN_TAB    lib.TransitionType = "OPEN_TAB"
	TR_CLOSE_TAB   lib.TransitionType = "CLOSE_TAB"
	TR_NEXT_TAB    lib.TransitionType = "NEXT_TAB"
	TR_PREV_TAB    lib.TransitionType = "PREV_TAB"
	TR_CLOSE_APP   lib.TransitionType = "CLOSE_APP"

	STATE_WRITE_CMD  lib.StateType      = "WRITE_CMD"
	TR_START_WRITING lib.TransitionType = "START_WRITING"
	TR_CANCEL        lib.TransitionType = "CANCEL"
	TR_VALIDATE      lib.TransitionType = "VALIDATE"
)

type WindowMachineCtx struct {
	tabs        []*Tab
	selectedTab int
}

func buildWindowMachine() *lib.Machine {
	return lib.NewMachine(
		&WindowMachineCtx{
			tabs:        make([]*Tab, 0),
			selectedTab: 0,
		},
		STATE_SHOW_TAB,
		lib.States{
			STATE_SHOW_TAB: &lib.State{
				Transitions: lib.Transitions{
					TR_CLOSE_APP: &lib.Transition{
						Target: STATE_SHOW_TAB,
						Action: func(c interface{}, ev *lib.Event) {
							App.Stop()
						},
					},
					TR_OPEN_TAB: &lib.Transition{
						Target: STATE_SHOW_TAB,
						Action: func(c interface{}, ev *lib.Event) {
							wmc := c.(*WindowMachineCtx)
							newtab := ev.Payload.(*Tab)
							wmc.tabs = append(wmc.tabs, newtab)
							wmc.selectedTab = len(wmc.tabs) - 1
						},
					},
					TR_CLOSE_TAB:     &lib.Transition{
						Target: STATE_SHOW_TAB,
						Action: func(c interface{}, ev *lib.Event) {
							wmc := c.(*WindowMachineCtx)
							if len(wmc.tabs) == 1 {
								return
							}
							i := wmc.selectedTab
							wmc.tabs = append(wmc.tabs[:i], wmc.tabs[i+1:]...)
							wmc.selectedTab = 0
						},
					},
					TR_NEXT_TAB:      &lib.Transition{
						Target: STATE_SHOW_TAB,
						Action: func(c interface{}, ev *lib.Event) {
							wmc := c.(*WindowMachineCtx)
							next := wmc.selectedTab + 1
							if next >= len(wmc.tabs) {
								next = 0
							}
							wmc.selectedTab = next
						},
					},
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
	bindings config.Keybindings

	triggerRedraw atomic.Value // bool
}

func NewWindow(cfg []*config.Account, bindings config.Keybindings) *Window {
	w := &Window{
		machine: buildWindowMachine(),
		bindings: bindings,
		ex:      NewStatus("ici c'est pour les commandes"),
	}
	w.machine.OnTransition(func(s *lib.State, ev *lib.Event) {
		switch ev.Transition {
		case TR_OPEN_TAB:
			w.onOpenTab(s, ev)
		case TR_NEXT_TAB, TR_CLOSE_TAB:
			w.screen.Clear()
			w.AskRedraw()
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
		accwidget := NewMailboxesView(c.Name, w.bindings[config.KEY_MODE_MBOXES], w.onSelectMailbox)
		App.PostMessage(
			&workers.FetchMailboxes{},
			c.Name,
			func(response workers.Message) error {
				App.logger.Debug("fetchmailboxes callback")
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
		w.machine.Send(&lib.Event{TR_OPEN_TAB, &Tab{accwidget, c.Name, nil}})
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
	w.machine.Send(&lib.Event{TR_OPEN_TAB, &Tab{mv, mailbox.TabTitle(), nil}})
}

func (w *Window) onOpenTab(s *lib.State, ev *lib.Event) {
	tab := ev.Payload.(*Tab)
	tab.Content.AskingRedraw(func() {
		w.AskRedraw()
	})
	if w.screen != nil {
		w.screen.Clear()
		tab.Content.SetViewPort(w.tabViewPort())
	}
}

func (w *Window) tabViewPort() *views.ViewPort {
	_, h := w.screen.Size()
	return views.NewViewPort(w.screen, 0, 2, -1, h-1)
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
		t.Content.SetViewPort(w.tabViewPort())
	}
}
func (w *Window) Draw() {
	width, _ := w.screen.Size()
	styleBase := tcell.StyleDefault
	styleRev := styleBase.Reverse(true)
	for x := 0 ; x <= width ; x++ {
		w.screen.SetContent(x, 0, ' ', nil, styleBase)
		w.screen.SetContent(x, 1, '─', nil, styleBase)
	}
	s := w.state()
	offset := 1
	for i, t := range s.tabs {
		style := styleBase
		if i == s.selectedTab {
			style = styleRev
		}
		for x, runec := range t.Title {
			w.screen.SetContent(offset + x, 0, runec, nil, style)
		}
		offset += len(t.Title) + 1
	}
	s.tabs[s.selectedTab].Content.Draw()
	w.ex.Draw()
}

func (w *Window) HandleEvent(ev tcell.Event) bool {
	ks := []*lib.KeyStroke{lib.KeyStrokeFromEvent(ev)}
	if w.machine.Current == STATE_WRITE_CMD {
		// in write_cmd state always forward event to ex
		return w.ex.HandleEvent(ks)
	} else {
		// first check if this event refer toa global command
		if cmd := w.bindings[config.KEY_MODE_GLOBAL].FindCommand(ks); cmd != "" {
			// this is global command, so window should try to handle it
			machineEv, err := w.machine.BuildEvent(cmd)
			if err != nil || machineEv == nil{
				App.logger.Errorf("error building machine event from `%s` (%v)", cmd, err)
				return false
			}
			App.logger.Debugf("machineEvent %#v", machineEv)
			if w.machine.Send(machineEv) {
				return true
			}
		}
		// either not a global command or this tcell event does not translate
		// to an application machine event
		s := w.state()
		return s.tabs[s.selectedTab].Content.HandleEvent(ks)
	}
	return false
}
