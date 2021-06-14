package statesmachines

import (
	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/models"
)

const (
	STATE_SHOW_MAIL  lib.StateType      = "SHOW_MAIL"
	// TR_UP_MAIL        lib.TransitionType = "UP_MAIL"
	// TR_DOWN_MAIL      lib.TransitionType = "DOWN_MAIL"
	// TR_SET_MAILS      lib.TransitionType = "SET_MAILS"
)

type MailMachineCtx struct {
	Mail    *models.Mail
}

func NewMailMachine() *lib.Machine {
	return lib.NewMachine(
        &MailMachineCtx{},
        STATE_SHOW_MAIL,
        lib.States{
            STATE_SHOW_MAIL: &lib.State{
                Transitions: lib.Transitions{},
            },
        },
    )
}
