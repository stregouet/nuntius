package ui

import (
	"github.com/gdamore/tcell/v2"

    "github.com/stregouet/nuntius/models"
    "github.com/stregouet/nuntius/widgets"
)

type MailboxesView struct {
    mailboxes []*models.Mailbox
	accountName string
    *widgets.ListWidget
}

func NewMailboxesView(accountName string) *MailboxesView {
    l := widgets.NewList()
    return &MailboxesView{
		accountName: accountName,
        ListWidget: l,
    }
}

func (mv *MailboxesView) SetMailboxes(mboxes []*models.Mailbox) {
	mv.mailboxes = mboxes
    for _, mbox := range mboxes {
        mv.AddLine(mbox)
    }
	mv.AskRedraw()
}

func (mv *MailboxesView) HandleEvent(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyRune:
			switch ev.Rune() {
			case 'Q', 'q':
                App.Stop()
				return
			}
		}
	}
    mv.ListWidget.HandleEvent(ev)
}
