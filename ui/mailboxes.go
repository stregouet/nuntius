package ui

import (
	"github.com/gdamore/tcell/v2"

	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/models"
	"github.com/stregouet/nuntius/widgets"
)

const (
	STATE_LOAD_MBOXES lib.StateType      = "LOAD_MBOXES"
	STATE_SHOW_MBOXES lib.StateType      = "SHOW_MBOXES"
	TR_NEXT_MBOX      lib.TransitionType = "NEXT_MBOX"
	TR_PREV_MBOX      lib.TransitionType = "PREV_MBOX"
	TR_SET_MBOXES     lib.TransitionType = "SET_MBOXES"
)

type MailboxesMachineCtx struct {
	mboxes   []*models.Mailbox
	selected int
}

func buildMailboxesMachine() *lib.Machine {
	setmboxes := &lib.Transition{
		Target: STATE_SHOW_MBOXES,
		Action: func(c interface{}, ev *lib.Event) {
			state := c.(*MailboxesMachineCtx)
			mboxes := ev.Payload.([]*models.Mailbox)
			state.mboxes = mboxes
		},
	}
	return lib.NewMachine(
		&MailboxesMachineCtx{
			mboxes:   make([]*models.Mailbox, 0),
			selected: 0,
		},
		STATE_LOAD_MBOXES,
		lib.States{
			STATE_LOAD_MBOXES: &lib.State{
				Transitions: lib.Transitions{
					TR_SET_MBOXES: setmboxes,
				},
			},
			STATE_SHOW_MBOXES: &lib.State{
				Transitions: lib.Transitions{
					TR_SET_MBOXES: setmboxes,
					TR_NEXT_MBOX: &lib.Transition{
						Target: STATE_SHOW_MBOXES,
						Action: func(c interface{}, ev *lib.Event) {
							state := c.(*MailboxesMachineCtx)
							state.selected++
						},
					},
					TR_PREV_MBOX: &lib.Transition{
						Target: STATE_SHOW_MBOXES,
						Action: func(c interface{}, ev *lib.Event) {
							state := c.(*MailboxesMachineCtx)
							state.selected--
						},
					},
				},
			},
			STATE_WRITE_CMD: &lib.State{
				Transitions: lib.Transitions{
					TR_CANCEL:   &lib.Transition{Target: STATE_SHOW_TAB},
					TR_VALIDATE: &lib.Transition{Target: STATE_SHOW_TAB},
				},
			},
		},
	)
}

type MailboxesView struct {
	machine     *lib.Machine
	accountName string
	*widgets.TreeWidget
}

func NewMailboxesView(accountName string, onSelect func(accname string, m *models.Mailbox)) *MailboxesView {
	t := widgets.NewTree()
	t.OnSelect = func(line widgets.ITreeLine) {
		m := line.(*models.Mailbox)
		onSelect(accountName, m)
	}
	return &MailboxesView{
		machine:     buildMailboxesMachine(),
		accountName: accountName,
		TreeWidget:  t,
	}
}
func (mv *MailboxesView) Draw() {
	if mv.machine.Current == STATE_LOAD_MBOXES {
		style := tcell.StyleDefault
		for i, c := range "loading..." {
			mv.SetContent(i, 0, c, nil, style)
		}
	} else {
		mv.TreeWidget.Draw()
	}
}

func (mv *MailboxesView) SetMailboxes(mboxes []*models.Mailbox) {
	mv.machine.Send(&lib.Event{TR_SET_MBOXES, mboxes})
	mv.ClearLines()
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
	mv.TreeWidget.HandleEvent(ev)
}
