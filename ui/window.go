package ui

import (
	"fmt"
	// "os/exec"
	"sync/atomic"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"

	"github.com/stregouet/nuntius/config"
	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/models"
	sm "github.com/stregouet/nuntius/statesmachines"
	"github.com/stregouet/nuntius/workers"
	// "github.com/stregouet/nuntius/widgets"
)

type Window struct {
	screen tcell.Screen

	machine  *lib.Machine
	ex       *Status
	bindings config.Keybindings
	filters  config.Filters

	triggerRedraw atomic.Value // bool
}

func NewWindow(cfg *config.Config) *Window {
	w := &Window{
		machine:  sm.NewWindowMachine(),
		bindings: cfg.Keybindings,
		ex:       NewStatus("ici c'est pour les commandes"),
		filters:  cfg.Filters,
	}
	w.machine.OnTransition(func(s lib.StateType, ctx interface{}, ev *lib.Event) {
		if ev.Transition == sm.TR_CLOSE_APP {
			App.Stop()
			return
		}
		switch ev.Transition {
		case sm.TR_COMPOSE_MAIL:
			// XXX it should be possible to choose account user want to send mail with
			c := NewComposeView(cfg.Accounts[0], w.bindings[config.KEY_MODE_COMPOSE], w.Errorf)
			w.machine.Send(&lib.Event{sm.TR_OPEN_TAB, &sm.Tab{c, "compose"}})
		case sm.TR_OPEN_TAB:
			w.onOpenTab(ev)
			w.AskRedraw()
		case sm.TR_NEXT_TAB, sm.TR_PREV_TAB, sm.TR_CLOSE_TAB:
			w.AskRedraw()
		}
	})
	w.ex.AskingRedraw(func() {
		w.AskRedraw()
	})
	w.ResetRedraw()

	for _, c := range cfg.Accounts {
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
				switch r := response.(type) {
				case *workers.Error:
					App.logger.Errorf("fetchmailboxes %v", response)
					w.ShowMessage(r.Error.Error())
				case *workers.FetchMailboxesRes:
					accwidget.SetMailboxes(r.Mailboxes)
				default:
					App.logger.Error("unknown response type")
				}
				return nil
			})
		w.machine.Send(&lib.Event{sm.TR_OPEN_TAB, &sm.Tab{accwidget, c.Name}})
	}

	return w
}

func (w *Window) state() *sm.WindowMachineCtx {
	return w.machine.Context.(*sm.WindowMachineCtx)
}

func (w *Window) onSelectMailbox(acc string, mailbox *models.Mailbox) {
	mv := NewMailboxView(acc, mailbox.Name, w.bindings[config.KEY_MODE_MBOX], w.onSelectThread)
	App.PostDbMessage(
		&workers.FetchMailbox{Mailbox: mailbox.Name},
		acc,
		func(response workers.Message) error {
			switch r := response.(type) {
			case *workers.Error:
				App.logger.Errorf("fetchmailbox res %v", response)
				w.ShowMessage(r.Error.Error())
			case *workers.FetchMailboxRes:
				mv.SetThreads(r.List)
				mv.FetchNewMessages(r.LastSeenUid)
				mv.FetchUpdateMessages(r.LastSeenUid)
			default:
				App.logger.Error("unknown response type")
			}
			return nil
		})
	w.machine.Send(&lib.Event{sm.TR_OPEN_TAB, &sm.Tab{mv, mailbox.TabTitle()}})
}

func (w *Window) onSelectThread(acc, mailbox string, thread *models.Thread) {
	tv := NewThreadView(acc, mailbox, w.bindings[config.KEY_MODE_THREAD], w.onSelectMail)
	App.PostDbMessage(
		&workers.FetchThread{RootId: thread.RootId},
		acc,
		func(response workers.Message) error {
			switch r := response.(type) {
			case *workers.Error:
				w.ShowMessage(r.Error.Error())
			case *workers.FetchThreadRes:
				tv.SetMails(r.Mails)
			default:
				App.logger.Error("unknown response type")
			}
			return nil
		})
	w.machine.Send(&lib.Event{sm.TR_OPEN_TAB, &sm.Tab{tv, thread.Subject}})
}

func (w *Window) onSelectMail(acc, mailbox string, mail *models.Mail) {
	mv := NewMailView(w.bindings[config.KEY_MODE_MAIL], w.bindings[config.KEY_MODE_PARTS], w.filters, mail)
	mv.OnSetViewPort(func(view *views.ViewPort, screen tcell.Screen) {
		mv.SetPartsView(view)
	})
	App.PostImapMessage(
		&workers.FetchFullMail{Uid: mail.Uid, Mailbox: mailbox},
		acc,
		func(response workers.Message) error {
			switch r := response.(type) {
			case *workers.Error:
				w.ShowMessage(r.Error.Error())
			case *workers.FetchFullMailRes:
				mv.SetFilepath(r.Filepath)
				App.logger.Debugf("full mail received, `%v`", r.Filepath)
			default:
				App.logger.Error("unknown response type")
			}
			return nil
		})
	w.machine.Send(&lib.Event{sm.TR_OPEN_TAB, &sm.Tab{mv, mail.Subject}})
}

func (w *Window) onOpenTab(ev *lib.Event) {
	tab := ev.Payload.(*sm.Tab)
	tab.Content.AskingRedraw(func() {
		w.AskRedraw()
	})
	if w.screen != nil {
		w.screen.Clear()
		tab.Content.SetViewPort(w.tabViewPort(), w.screen)
	}
}

func (w *Window) tabViewPort() *views.ViewPort {
	_, h := w.screen.Size()
	return views.NewViewPort(w.screen, 0, 2, -1, h-3)
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

func (w *Window) Errorf(msg string, args ...interface{}) {
	m := fmt.Sprintf(msg, args...)
	w.ex.ShowError(m)
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
	w.ex.SetViewPort(w.exViewPort(), s)
	state := w.state()
	for _, t := range state.Tabs {
		t.Content.SetViewPort(w.tabViewPort(), s)
	}
}
func (w *Window) Draw() {
	w.screen.HideCursor()
	width, _ := w.screen.Size()
	styleBase := tcell.StyleDefault
	styleRev := styleBase.Reverse(true)
	for x := 0; x <= width; x++ {
		w.screen.SetContent(x, 0, ' ', nil, styleBase)
		w.screen.SetContent(x, 1, '─', nil, styleBase)
	}
	s := w.state()
	offset := 1
	for i, t := range s.Tabs {
		style := styleBase
		if i == s.SelectedTab {
			style = styleRev
		}
		for x, runec := range t.Title {
			w.screen.SetContent(offset+x, 0, runec, nil, style)
		}
		offset += len(t.Title) + 1
	}
	s.Tabs[s.SelectedTab].Content.Draw()
	w.ex.Draw()
}

func (w *Window) HandleEvent(ev tcell.Event) bool {
	ks := []*lib.KeyStroke{lib.KeyStrokeFromEvent(ev)}
	if w.machine.Current == sm.STATE_WRITE_CMD {
		// in write_cmd state always forward event to ex
		return w.ex.HandleEvent(ks)
	} else {
		s := w.state()
		curTab := s.Tabs[s.SelectedTab]
		if curTab.Content.IsActiveTerm() {
			return curTab.Content.HandleEvent(ks)
		}
		// first check if this event refer toa global command
		if cmd := w.bindings[config.KEY_MODE_GLOBAL].FindCommand(ks); cmd != "" {
			// this is global command, so window should try to handle it
			machineEv, err := w.machine.BuildEvent(cmd)
			if err != nil || machineEv == nil {
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
		return curTab.Content.HandleEvent(ks)
	}
	return false
}

func (w *Window) HandleTransitions(ev *lib.Event) {
	s := w.state()
	if w.ex.HandleTransitions(ev) {
		return
	}
	for _, t := range s.Tabs {
		if t.Content.HandleTransitions(ev) {
			return
		}
	}

}
