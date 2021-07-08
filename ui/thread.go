package ui

import (
	"github.com/stregouet/nuntius/config"
	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/models"
	sm "github.com/stregouet/nuntius/statesmachines"
	"github.com/stregouet/nuntius/widgets"
)

type ThreadView struct {
	machine  *lib.Machine
	bindings config.Mapping
	subject  string
	*widgets.TreeWidget
}

func NewThreadView(accname, mailbox, subject string, bindings config.Mapping, onSelect func(accname, mailbox string, m *models.Mail)) *ThreadView {
	t := widgets.NewTree()
	machine := sm.NewThreadMachine()
	machine.OnTransition(func(s lib.StateType, ctx interface{}, ev *lib.Event) {
		state := ctx.(*sm.ThreadMachineCtx)
		switch ev.Transition {
		case sm.TR_SELECT_MAIL:
			onSelect(accname, mailbox, state.Mails[state.Selected-1])
		case sm.TR_UP_MAIL, sm.TR_DOWN_MAIL:
			t.SetSelected(state.Selected)
		}
	})
	return &ThreadView{
		machine:    machine,
		bindings:   bindings,
		subject:    subject,
		TreeWidget: t,
	}
}

// Tab interface
func (tv *ThreadView) TabTitle() string {
	return "\uf086 " + tv.thread.Subject
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
