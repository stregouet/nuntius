package ui

import (
	"github.com/stregouet/nuntius/config"
	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/models"
	sm "github.com/stregouet/nuntius/statesmachines"
	"github.com/stregouet/nuntius/widgets"
)

type MailPartsView struct {
	machine *lib.Machine
	bindings config.Mapping
	*widgets.TreeWidget
}

func NewMailPartsView(bindings config.Mapping, parts []*models.BodyPart, onSelect func(part models.BodyPath)) *MailPartsView {
	machine := sm.NewMailPartsMachine(parts)
	mp := &MailPartsView{machine, bindings, nil}
	t := widgets.NewTreeWithInitSelected(mp.state().Selected)
	machine.OnTransition(func(s lib.StateType, ctx interface{}, ev *lib.Event) {
		state := ctx.(*sm.MailPartsMachineCtx)
		switch ev.Transition {
		case sm.TR_SELECT_PART:
			onSelect(state.Parts[state.Selected-1].Path)
		case sm.TR_MAIL_PARTS_UP, sm.TR_MAIL_PARTS_DOWN, sm.TR_SET_SELECTED_PART:
			t.SetSelected(state.Selected)
			t.AskRedraw()
		}
	})
	for _, part := range parts {
		t.AddLine(part)
	}
	mp.TreeWidget = t
	return mp
}

func (mp *MailPartsView) state() *sm.MailPartsMachineCtx {
	return mp.machine.Context.(*sm.MailPartsMachineCtx)
}

func (mp *MailPartsView) HandleEvent(ks []*lib.KeyStroke) bool {
	if cmd := mp.bindings.FindCommand(ks); cmd != "" {
		mev, err := mp.machine.BuildEvent(cmd)
		if err != nil {
			App.logger.Errorf("error building machine event from `%s` (%v)", cmd, err)
			return false
		}
		if mp.machine.Send(mev) {
			return true
		}
	}
	return false
}
