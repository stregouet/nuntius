package statesmachines

import (
	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/models"
)

const (
	STATE_LOAD_MBOXES lib.StateType      = "LOAD_MBOXES"
	STATE_SHOW_MBOXES lib.StateType      = "SHOW_MBOXES"
	TR_UP_MBOX        lib.TransitionType = "UP_MBOX"
	TR_DOWN_MBOX      lib.TransitionType = "DOWN_MBOX"
	TR_SET_MBOXES     lib.TransitionType = "SET_MBOXES"
	TR_SELECT_MBOX    lib.TransitionType = "SELECT_MBOX"
)

type MailboxesMachineCtx struct {
	Mboxes   []*models.Mailbox
	Selected int
}

func NewMailboxesMachine() *lib.Machine {
	setmboxes := &lib.Transition{
		Target: STATE_SHOW_MBOXES,
		Action: func(c interface{}, ev *lib.Event) {
			state := c.(*MailboxesMachineCtx)
			Mboxes := ev.Payload.([]*models.Mailbox)
			state.Mboxes = Mboxes
		},
	}
	return lib.NewMachine(
		&MailboxesMachineCtx{
			Mboxes:   make([]*models.Mailbox, 0),
			Selected: 1,
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
					TR_SELECT_MBOX: &lib.Transition{
						Target: STATE_SHOW_MBOXES,
					},
					TR_DOWN_MBOX: &lib.Transition{
						Target: STATE_SHOW_MBOXES,
						Action: func(c interface{}, ev *lib.Event) {
							state := c.(*MailboxesMachineCtx)
							next := state.Selected + 1
							if next > len(state.Mboxes) {
								next = len(state.Mboxes)
							}
							state.Selected = next
						},
					},
					TR_UP_MBOX: &lib.Transition{
						Target: STATE_SHOW_MBOXES,
						Action: func(c interface{}, ev *lib.Event) {
							state := c.(*MailboxesMachineCtx)
							next := state.Selected - 1
							if next < 1 {
								next = 1
							}
							state.Selected = next
						},
					},
				},
			},
		},
	)
}
