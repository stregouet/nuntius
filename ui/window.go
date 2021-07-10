package ui

import (
	"fmt"
	// "os/exec"
	"sync/atomic"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
	"github.com/mattn/go-runewidth"

	"github.com/stregouet/nuntius/config"
	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/models"
	sm "github.com/stregouet/nuntius/statesmachines"
	"github.com/stregouet/nuntius/workers"
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
		filters:  cfg.Filters,
	}
	w.ex = NewStatus("ici c'est pour les commandes", w.OnExCmd)
	w.machine.OnTransition(func(s lib.StateType, ctx interface{}, ev *lib.Event) {
		if ev.Transition == sm.TR_CLOSE_APP {
			App.Stop()
			return
		}
		switch ev.Transition {
		case sm.TR_START_WRITING:
			w.ex.machine.Send(&lib.Event{sm.TR_STATUS_START_WRITING, nil})
		case sm.TR_COMPOSE_MAIL:
			// XXX it should be possible to choose account user want to send mail with
			c := NewComposeView(cfg.Accounts[0], w.bindings[config.KEY_MODE_COMPOSE])
			w.addTab(c)
		case sm.TR_OPEN_TAB:
			w.onOpenTab(ev)
			w.AskRedraw()
		case sm.TR_NEXT_TAB, sm.TR_PREV_TAB, sm.TR_CLOSE_TAB, sm.TR_END_CMD:
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
		w.addTab(accwidget)
	}

	return w
}

func (w *Window) OnExCmd(cmd string) {
	w.machine.Send(&lib.Event{sm.TR_END_CMD, nil})
	if cmd != "" {
		command, err := lib.ParseCmd(cmd)
		if err != nil {
			w.Errorf("cannot apply cmd %v", err)
			return
		}
		ev := &lib.Event{command.ToTrType(), command.Args}
		res := w.HandleTransitions(ev)
		if !res {
			w.ShowMessagef("no available command for `%s`", command.Name)
		}
	}
}

func (w *Window) addTab(content sm.Tab) {
	w.machine.Send(&lib.Event{sm.TR_OPEN_TAB, content})
}

func (w *Window) state() *sm.WindowMachineCtx {
	return w.machine.Context.(*sm.WindowMachineCtx)
}

func (w *Window) onSelectMailbox(acc string, mailbox *models.Mailbox) {
	mv := NewMailboxView(acc, mailbox, w.bindings[config.KEY_MODE_MBOX], w.onSelectThread)
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
	w.addTab(mv)
}

func (w *Window) onSelectThread(acc, mailbox string, thread *models.Thread) {
	var tab sm.Tab
	if thread.Count == 1 {
		tab = w.buildMailView(thread)
	} else {
		tab = NewThreadView(acc, mailbox, thread, w.bindings[config.KEY_MODE_THREAD], w.onSelectMail)
	}
	App.PostDbMessage(
		&workers.FetchThread{RootId: thread.RootId},
		acc,
		func(response workers.Message) error {
			switch r := response.(type) {
			case *workers.Error:
				w.ShowMessage(r.Error.Error())
			case *workers.FetchThreadRes:
				switch t := tab.(type) {
				case *MailView:
					t.SetMail(r.Mails[0], mailbox, acc)
				case *ThreadView:
					t.SetMails(r.Mails)
				}
			default:
				App.logger.Error("unknown response type")
			}
			return nil
		})
	w.addTab(tab)
}

func (w *Window) buildMailView(thread *models.Thread) *MailView {
	mv := NewMailView(w.bindings[config.KEY_MODE_MAIL], w.bindings[config.KEY_MODE_PARTS], w.filters)
	mv.OnRead(func() {
		App.logger.Debugf("one mail marked as read %d", thread.SeenCount)
		thread.MarkOneAsRead()
	})
	return mv
}

func (w *Window) onSelectMail(acc, mailbox string, mail *models.Mail, thread *models.Thread) {
	mv := w.buildMailView(thread)
	mv.SetMail(mail, mailbox, acc)
	w.addTab(mv)
}

func (w *Window) onOpenTab(ev *lib.Event) {
	tab := ev.Payload.(sm.Tab)
	tab.AskingRedraw(func() {
		w.AskRedraw()
	})
	if w.screen != nil {
		w.screen.Clear()
		tab.SetViewPort(w.tabViewPort(), w.screen)
	}
	tab.OnMessage(w.Errorf)
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
		t.SetViewPort(w.tabViewPort(), s)
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

		title := runewidth.Truncate(t.TabTitle(), 12, "…")
		for x, runec := range []rune(title) {
			w.screen.SetContent(offset+x, 0, runec, nil, style)
		}
		offset += runewidth.StringWidth(title) + 1
	}
	s.Tabs[s.SelectedTab].Draw()
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
		if curTab.IsActiveTerm() {
			return curTab.HandleEvent(ks)
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
		return curTab.HandleEvent(ks)
	}
	return false
}

func (w *Window) HandleTransitions(ev *lib.Event) bool {
	s := w.state()
	if w.ex.HandleTransitions(ev) {
		return true
	}
	for _, t := range s.Tabs {
		if t.HandleTransitions(ev) {
			return true
		}
	}
	return false
}
