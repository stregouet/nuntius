package statesmachines

import (
	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/models"
)

const (
	STATE_LOAD_MBOX lib.StateType      = "LOAD_MBOX"
	STATE_SHOW_MBOX lib.StateType      = "SHOW_MBOX"
	TR_SET_THREADS  lib.TransitionType = "SET_THREADS"
	TR_REFRESH_MBOX lib.TransitionType = "REFRESH_MBOX"
)

type MailboxMachineCtx struct {
	threads  []*models.Thread
	selected int
}

func NewMailboxMachine() *lib.Machine {
	c := &MailboxMachineCtx{
		threads:  make([]*models.Thread, 0),
		selected: 0,
	}
	return lib.NewMachine(
		c,
		STATE_LOAD_MBOX,
		lib.States{
			STATE_SHOW_MBOX: &lib.State{
				Transitions: lib.Transitions{
					TR_REFRESH_MBOX: &lib.Transition{
						Target: STATE_LOAD_MBOX,
					},
				},
			},
			STATE_LOAD_MBOX: &lib.State{
				Transitions: lib.Transitions{
					TR_SET_THREADS: &lib.Transition{
						Target: STATE_SHOW_MBOX,
						Action: func(c interface{}, ev *lib.Event) {
							state := c.(*MailboxMachineCtx)
							threads := ev.Payload.([]*models.Thread)
							state.threads = threads
						},
					},
				},
			},
		},
	)
}
