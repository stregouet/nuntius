package statesmachines

import (
	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/models"
)

const (
	STATE_LOAD_MAIL     lib.StateType      = "LOAD_MAIL"
	STATE_SHOW_MAIL     lib.StateType      = "SHOW_MAIL"
	STATE_SHOW_MAIL_PARTS lib.StateType      = "SHOW_MAIL_PARTS"
	TR_SCROLL_UP_MAIL   lib.TransitionType = "SCROLL_UP_MAIL"
	TR_SCROLL_DOWN_MAIL lib.TransitionType = "SCROLL_DOWN_MAIL"
	TR_SET_FILEPATH     lib.TransitionType = "SET_FILEPATH"
	TR_SHOW_MAIL_PARTS  lib.TransitionType = "SHOW_MAIL_PARTS"
	TR_SHOW_MAIL_PART   lib.TransitionType = "SHOW_MAIL_PART"
	// TR_DOWN_MAIL      lib.TransitionType = "DOWN_MAIL"
	// TR_SET_MAILS      lib.TransitionType = "SET_MAILS"
)

type MailMachineCtx struct {
	Mail         *models.Mail
	Filepath     string
	SelectedPart *models.BodyPath
}

func NewMailMachine(m *models.Mail) *lib.Machine {
	path := m.FindPlaintext()
	if path == nil {
		path = m.FindFirstNonMultipart()
	}
	return lib.NewMachine(
		&MailMachineCtx{m, "", path},
		STATE_LOAD_MAIL,
		lib.States{
			STATE_SHOW_MAIL_PARTS: &lib.State{
				Transitions: lib.Transitions{
					TR_SHOW_MAIL_PART: &lib.Transition{
						Target: STATE_SHOW_MAIL,
						Action: func(c interface{}, ev *lib.Event) {
							state := c.(*MailMachineCtx)
							bp := ev.Payload.(models.BodyPath)
							state.SelectedPart = &bp
						},
					},
				},
			},
			STATE_LOAD_MAIL: &lib.State{
				Transitions: lib.Transitions{
					TR_SET_FILEPATH: &lib.Transition{
						Target: STATE_SHOW_MAIL,
						Action: func(c interface{}, ev *lib.Event) {
							state := c.(*MailMachineCtx)
							filepath := ev.Payload.(string)
							state.Filepath = filepath
						},
					},
				},
			},
			STATE_SHOW_MAIL: &lib.State{
				Transitions: lib.Transitions{
					TR_SCROLL_UP_MAIL: &lib.Transition{
						Target: STATE_SHOW_MAIL,
					},
					TR_SCROLL_DOWN_MAIL: &lib.Transition{
						Target: STATE_SHOW_MAIL,
					},
					TR_SHOW_MAIL_PARTS: &lib.Transition{
						Target: STATE_SHOW_MAIL_PARTS,
					},
				},
			},
		},
	)
}
