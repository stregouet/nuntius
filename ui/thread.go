package ui

import (
	"github.com/stregouet/nuntius/config"
	"github.com/stregouet/nuntius/models"
	"github.com/stregouet/nuntius/lib"
	sm "github.com/stregouet/nuntius/statesmachines"
	"github.com/stregouet/nuntius/widgets"
)

type ThreadView struct {
	machine  *lib.Machine
	bindings config.Mapping
	*widgets.TreeWidget
}

func NewThreadView(accname string, bindings config.Mapping, onSelect func(accname string, m *models.Mail)) *ThreadView {
	t := widgets.NewTree()
	machine := sm.NewThreadMachine()
	machine.OnTransition(func(s lib.StateType, ctx interface{}, ev *lib.Event) {
		state := ctx.(*sm.ThreadMachineCtx)
		switch ev.Transition {
		case sm.TR_SELECT_MAIL:
			onSelect(accname, state.Mails[state.Selected-1])
		case sm.TR_UP_MAIL, sm.TR_DOWN_MAIL:
			t.SetSelected(state.Selected)
		}
	})
	return &ThreadView{
		machine:    machine,
		bindings:   bindings,
		TreeWidget: t,
	}
}

func (tv *ThreadView) SetMails(mails []*models.Mail) {
	tv.machine.Send(&lib.Event{sm.TR_SET_MAILS, mails})
	tv.ClearLines()
	for _, mail := range mails {
		tv.AddLine(mail)
	}
	tv.AskRedraw()
}

func (tv *ThreadView) HandleEvent(ks []*lib.KeyStroke) bool {
	if cmd := tv.bindings.FindCommand(ks); cmd != "" {
		mev, err := tv.machine.BuildEvent(cmd)
		if err != nil {
			App.logger.Errorf("error building machine event from `%s` (%v)", cmd, err)
			return false
		}
		if tv.machine.Send(mev) {
			return true
		}
	}
	return false
}
