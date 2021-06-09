package ui

import (
	"github.com/gdamore/tcell/v2"

	"github.com/stregouet/nuntius/config"
	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/models"
	sm "github.com/stregouet/nuntius/statesmachines"
	"github.com/stregouet/nuntius/widgets"
)

type MailboxesView struct {
	machine     *lib.Machine
	accountName string
	bindings    config.Mapping
	*widgets.TreeWidget
}

func NewMailboxesView(accountName string, bindings config.Mapping, onSelect func(accname string, m *models.Mailbox)) *MailboxesView {
	t := widgets.NewTree()
	t.OnSelect = func(line widgets.ITreeLine) {
		m := line.(*models.Mailbox)
		onSelect(accountName, m)
	}
	machine := sm.NewMailboxesMachine()
	machine.OnTransition(func(s lib.StateType, ctx interface{}, ev *lib.Event) {
		state := ctx.(*sm.MailboxesMachineCtx)
		switch ev.Transition {
		case sm.TR_SELECT_MBOX:
			onSelect(accountName, state.Mboxes[state.Selected - 1])
		case sm.TR_UP_MBOX, sm.TR_DOWN_MBOX:
			t.SetSelected(state.Selected)
		}
	})
	return &MailboxesView{
		machine:     machine,
		accountName: accountName,
		bindings:    bindings,
		TreeWidget:  t,
	}
}
func (mv *MailboxesView) Draw() {
	if mv.machine.Current == sm.STATE_LOAD_MBOXES {
		style := tcell.StyleDefault
		for i, c := range "loading..." {
			mv.SetContent(i, 0, c, nil, style)
		}
	} else {
		mv.TreeWidget.Draw()
	}
}

func (mv *MailboxesView) SetMailboxes(mboxes []*models.Mailbox) {
	mv.machine.Send(&lib.Event{sm.TR_SET_MBOXES, mboxes})
	mv.ClearLines()
	for _, mbox := range mboxes {
		mv.AddLine(mbox)
	}
	mv.AskRedraw()
}

func (mv *MailboxesView) HandleEvent(ks []*lib.KeyStroke) bool {
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
