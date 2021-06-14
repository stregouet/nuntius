package statesmachines

import (
	"strconv"

	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/models"
)

const (
	STATE_LOAD_MBOX lib.StateType      = "LOAD_MBOX"
	STATE_SHOW_MBOX lib.StateType      = "SHOW_MBOX"
	TR_SET_THREADS  lib.TransitionType = "SET_THREADS"
	TR_REFRESH_MBOX lib.TransitionType = "REFRESH_MBOX"
	TR_UP_THREAD    lib.TransitionType = "UP_THREAD"
	TR_DOWN_THREAD  lib.TransitionType = "DOWN_THREAD"
)

type MailboxMachineCtx struct {
	Threads  []*models.Thread
	Selected int
}

func getNblines(ev *lib.Event) int {
	if args, ok := ev.Payload.(lib.CmdArgs); ok {
		if line, err := strconv.Atoi(args["line"]); err == nil {
			return line
		}
	}
	return 1
}

func NewMailboxMachine() *lib.Machine {
	c := &MailboxMachineCtx{
		Threads:  make([]*models.Thread, 0),
		Selected: 1,
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
					TR_UP_THREAD: &lib.Transition{
						Target: STATE_SHOW_MBOX,
						Action: func(c interface{}, ev *lib.Event) {
							state := c.(*MailboxMachineCtx)
							next := state.Selected - getNblines(ev)
							if next < 1 {
								next = 1
							}
							state.Selected = next
						},
					},
					TR_DOWN_THREAD: &lib.Transition{
						Target: STATE_SHOW_MBOX,
						Action: func(c interface{}, ev *lib.Event) {
							state := c.(*MailboxMachineCtx)
							next := state.Selected + getNblines(ev)
							if next > len(state.Threads) {
								next = len(state.Threads)
							}
							state.Selected = next
						},
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
							state.Threads = threads
						},
					},
				},
			},
		},
	)
}
