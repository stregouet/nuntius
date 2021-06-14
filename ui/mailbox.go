package ui

import (
	"github.com/gdamore/tcell/v2"

	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/models"
	sm "github.com/stregouet/nuntius/statesmachines"
	"github.com/stregouet/nuntius/widgets"
	"github.com/stregouet/nuntius/workers"
)

type MailboxView struct {
	machine       *lib.Machine
	errorListener func(e error)
	accountName   string
	mboxName      string
	*widgets.ListWidget
}

func NewMailboxView(accountName, mboxName string, onSelect func(accname string, t *models.Thread)) *MailboxView {
	l := widgets.NewList()
	return &MailboxView{
		machine:     sm.NewMailboxMachine(),
		accountName: accountName,
		mboxName:    mboxName,
		ListWidget:  l,
	}
}

func (mv *MailboxView) SetThreads(threads []*models.Thread) {
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
	// mv.FetchUpdateMessages()
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
				mv.SetThreads(r.Threads)
			}
			return nil
		})
}

func (mv *MailboxView) Draw() {
	if mv.machine.Current == sm.STATE_LOAD_MBOX {
		style := tcell.StyleDefault
		for i, c := range "loading..." {
			mv.SetContent(i, 0, c, nil, style)
		}
	} else {
		mv.ListWidget.Draw()
	}
}

func (mv *MailboxView) HandleEvent(ks []*lib.KeyStroke) bool {
	return mv.ListWidget.HandleEvent(ks)
}
