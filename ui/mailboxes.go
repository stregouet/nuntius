package ui

import (
	"github.com/gdamore/tcell/v2"

    "github.com/stregouet/nuntius/models"
    "github.com/stregouet/nuntius/widgets"
)

type MailboxesView struct {
    mailboxes []*models.Mailbox
    *widgets.ListWidget
}

func NewMailboxesView(mboxes []string) *MailboxesView {
    l := widgets.NewList()
    mailboxes := make([]*models.Mailbox, 0, len(mboxes))
    for _, mbox := range mboxes {
        m := &models.Mailbox{mbox}
        mailboxes = append(mailboxes, m)
        l.AddLine(m)
    }
    return &MailboxesView{
        mailboxes: mailboxes,
        ListWidget: l,
    }
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
