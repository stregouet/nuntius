package ui

import (
	"github.com/gdamore/tcell/v2"

	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/models"
	sm "github.com/stregouet/nuntius/statesmachines"
	"github.com/stregouet/nuntius/widgets"
)

type MailboxView struct {
	machine *lib.Machine
	*widgets.ListWidget
}

func NewMailboxView(accountName string, onSelect func(accname string, t *models.Thread)) *MailboxView {
	l := widgets.NewList()
	return &MailboxView{
		machine:    sm.NewMailboxMachine(),
		ListWidget: l,
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
