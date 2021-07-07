package ui

import (
	"github.com/gdamore/tcell/v2"

	"github.com/stregouet/nuntius/config"
	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/models"
	sm "github.com/stregouet/nuntius/statesmachines"
	"github.com/stregouet/nuntius/widgets"
	"github.com/stregouet/nuntius/workers"
)

type MailboxView struct {
	machine           *lib.Machine
	viewableFirstLine int
	viewableLastLine  int
	bindings          config.Mapping
	errorListener     func(e error)
	accountName       string
	mboxName          string
	*widgets.ListWidget
}

func NewMailboxView(accountName, mboxName string, bindings config.Mapping, onSelect func(accname, mailbox string, t *models.Thread)) *MailboxView {
	machine := sm.NewMailboxMachine()
	l := widgets.NewList()
	machine.OnTransition(func(s lib.StateType, ctx interface{}, ev *lib.Event) {
		state := ctx.(*sm.MailboxMachineCtx)
		switch ev.Transition {
		case sm.TR_SELECT_THREAD:
			onSelect(accountName, mboxName, state.Threads[state.Selected-1])
		case sm.TR_UP_THREAD, sm.TR_DOWN_THREAD:
			l.SetSelected(state.Selected)
		}
	})
	return &MailboxView{
		machine:     machine,
		accountName: accountName,
		bindings:    bindings,
		mboxName:    mboxName,
		ListWidget:  l,
	}
}

func (mv *MailboxView) SetThreads(threads []*models.Thread) {
	if len(threads) == 0 {
		return
	}
	mv.machine.Send(&lib.Event{sm.TR_SET_THREADS, threads})
	mv.ClearLines()
	for _, t := range threads {
		mv.AddLine(t)
	}
	mv.AskRedraw()
}

func (mv *MailboxView) Error(err error) {
	if mv.errorListener != nil {
		mv.errorListener(err)
	}
}

func (mv *MailboxView) Refresh(lastuid uint32) {
	mv.FetchNewMessages(lastuid)
	mv.FetchUpdateMessages(lastuid)
}

func (mv *MailboxView) FetchUpdateMessages(lastuid uint32) {
	App.PostImapMessage(
		&workers.FetchMessageUpdates{Mailbox: mv.mboxName, LastSeenUid: lastuid},
		mv.accountName,
		func(response workers.Message) error {
			switch r := response.(type) {
			case *workers.Error:
				App.logger.Errorf("fetch update messages %v", response)
				mv.Error(r.Error)
			case *workers.FetchMessageUpdatesRes:
				mv.updateDb(r.Mails, lastuid)
			}
			return nil
		})
}

func (mv *MailboxView) FetchNewMessages(lastuid uint32) {
	App.PostImapMessage(
		&workers.FetchNewMessages{Mailbox: mv.mboxName, LastSeenUid: lastuid},
		mv.accountName,
		func(response workers.Message) error {
			switch r := response.(type) {
			case *workers.Error:
				App.logger.Errorf("fetch new message res %v", response)
				mv.Error(r.Error)
			case *workers.FetchNewMessagesRes:
				mv.insertDb(r.Mails)
			}
			return nil
		})
}

func (mv *MailboxView) updateDb(mails []*models.Mail, lastuid uint32) {
	if len(mails) == 0 {
		return
	}
	App.PostDbMessage(
		&workers.UpdateMessages{Mailbox: mv.mboxName, Mails: mails, LastSeenUid: lastuid},
		mv.accountName,
		func(response workers.Message) error {
			switch r := response.(type) {
			case *workers.Error:
				App.logger.Errorf("error updating mails flags in db %v", r)
				mv.Error(r.Error)
			case *workers.UpdateMessagesRes:
				mv.SetThreads(r.Threads)
			}
			return nil
		})
}

func (mv *MailboxView) insertDb(mails []*models.Mail) {
	if len(mails) == 0 {
		return
	}
	App.PostDbMessage(
		&workers.InsertNewMessages{Mailbox: mv.mboxName, Mails: mails},
		mv.accountName,
		func(response workers.Message) error {
			switch r := response.(type) {
			case *workers.Error:
				App.logger.Errorf("upsert messages res %v", response)
				mv.Error(r.Error)
			case *workers.InsertNewMessagesRes:
				App.logger.Debugf("correctly added %d in db", len(r.Threads))
				mv.SetThreads(r.Threads)
			}
			return nil
		})
}

func (mv *MailboxView) Draw() {
	mv.Clear()
	if mv.machine.Current == sm.STATE_LOAD_MBOX {
		style := tcell.StyleDefault
		mv.Print(0, 0, style, "loading...")
	} else {
		mv.ListWidget.Draw()
	}
}

func (mv *MailboxView) HandleEvent(ks []*lib.KeyStroke) bool {
	if cmd := mv.bindings.FindCommand(ks); cmd != "" {
		mev, err := mv.machine.BuildEvent(cmd)
		if err != nil {
			App.logger.Errorf("error building machine event from `%s` (%v)", cmd, err)
			return false
		}
		if mv.machine.Send(mev) {
			return true
		}
	}
	return false
}
