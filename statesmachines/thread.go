package statesmachines

import (
	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/models"
)

const (
	STATE_SHOW_THREAD lib.StateType      = "SHOW_THREAD"
	TR_UP_MAIL        lib.TransitionType = "UP_MAIL"
	TR_DOWN_MAIL      lib.TransitionType = "DOWN_MAIL"
	TR_SET_MAILS      lib.TransitionType = "SET_MAILS"
	TR_SELECT_MAIL    lib.TransitionType = "SELECT_MAIL"
)

type ThreadMachineCtx struct {
	Mails    []*models.Mail
	Selected int
}

func NewThreadMachine() *lib.Machine {
	return lib.NewMachine(
		&ThreadMachineCtx{
			Mails:    make([]*models.Mail, 0),
			Selected: 1,
		},
		STATE_SHOW_THREAD,
		lib.States{
			STATE_SHOW_THREAD: &lib.State{
				Transitions: lib.Transitions{
					TR_SELECT_MAIL: &lib.Transition{
						Target: STATE_SHOW_THREAD,
					},
					TR_SET_MAILS: &lib.Transition{
						Target: STATE_SHOW_THREAD,
						Action: func(c interface{}, ev *lib.Event) {
							state := c.(*ThreadMachineCtx)
							mails := ev.Payload.([]*models.Mail)
							state.Mails = mails
						},
					},
					TR_UP_MAIL: &lib.Transition{
						Target: STATE_SHOW_THREAD,
						Action: func(c interface{}, ev *lib.Event) {
							state := c.(*ThreadMachineCtx)
							next := state.Selected - 1
							if next < 1 {
								next = 1
							}
							state.Selected = next
						},
					},
					TR_DOWN_MAIL: &lib.Transition{
						Target: STATE_SHOW_THREAD,
						Action: func(c interface{}, ev *lib.Event) {
							state := c.(*ThreadMachineCtx)
							next := state.Selected + 1
							if next > len(state.Mails) {
								next = len(state.Mails)
							}
							state.Selected = next
						},
					},
				},
			},
		},
	)
}
