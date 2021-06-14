package ui


import (
	"github.com/gdamore/tcell/v2"

	"github.com/stregouet/nuntius/config"
	// "github.com/stregouet/nuntius/models"
	"github.com/stregouet/nuntius/lib"
	sm "github.com/stregouet/nuntius/statesmachines"
	"github.com/stregouet/nuntius/widgets"
)

type MailView struct {
	machine  *lib.Machine
	bindings config.Mapping
	*widgets.BaseWidget
}

func NewMailView(bindings config.Mapping) *MailView {
    b := widgets.BaseWidget{}
    machine := sm.NewMailMachine()
    return &MailView{
        machine: machine,
        BaseWidget: &b,
    }

}

func (mv *MailView) Draw() {
    style := tcell.StyleDefault
    for i, c := range "rien..." {
        mv.SetContent(i, 0, c, nil, style)
    }
}

func (mv *MailView) HandleEvent(ks []*lib.KeyStroke) bool {
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
