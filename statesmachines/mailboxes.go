package statesmachines

import (
	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/models"
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

func NewMailboxesMachine() *lib.Machine {
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
		},
	)
}
