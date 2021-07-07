package statesmachines

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/stregouet/nuntius/lib"
)

const (
	STATE_COMPOSE_WRITE_MAIL  lib.StateType = "WRITE_MAIL"
	STATE_COMPOSE_REVIEW_MAIL lib.StateType = "REVIEW_MAIL"
	STATE_COMPOSE_ERR         lib.StateType = "COMPOSE_ERR"

	TR_COMPOSE_REVIEW  lib.TransitionType = "COMPOSE_REVIEW"
	TR_COMPOSE_WRITE   lib.TransitionType = "COMPOSE_WRITE"
	TR_COMPOSE_SET_ERR lib.TransitionType = "COMPOSE_SET_ERR"
	TR_COMPOSE_SEND    lib.TransitionType = "COMPOSE_SEND"
)

type ComposeMachineCtx struct {
	MailFile *os.File
	Body     string
}

func NewComposeMachine(mailfile *os.File) *lib.Machine {
	writeTr := &lib.Transition{
		Target: STATE_COMPOSE_WRITE_MAIL,
		Action: func(c interface{}, ev *lib.Event) {
			state := c.(*ComposeMachineCtx)
			email, err := ioutil.TempFile("", "nuntius-*.eml")
			if err != nil {
				// XXX handle error
				return
			}
			_, err = email.WriteString(state.Body)
			if err != nil {
				// XXX handle error
				return
			}
			state.MailFile.Seek(0, io.SeekStart)
			state.MailFile = email
		},
	}

	return lib.NewMachine(
		&ComposeMachineCtx{mailfile, ""},
		STATE_COMPOSE_WRITE_MAIL,
		lib.States{
			STATE_COMPOSE_WRITE_MAIL: &lib.State{
				Transitions: lib.Transitions{
					TR_COMPOSE_REVIEW: &lib.Transition{
						Target: STATE_COMPOSE_REVIEW_MAIL,
						Action: func(c interface{}, ev *lib.Event) {
							state := c.(*ComposeMachineCtx)
							if state.MailFile == nil {
								return
							}
							defer func() {
								state.MailFile.Close()
								os.Remove(state.MailFile.Name())
								state.MailFile = nil
							}()
							state.MailFile.Sync()
							state.MailFile.Seek(0, io.SeekStart)
							content, err := io.ReadAll(state.MailFile)
							if err != nil {
								// XXX handle error
							}
							state.Body = string(content)
						},
					},
					TR_COMPOSE_SET_ERR: &lib.Transition{
						Target: STATE_COMPOSE_ERR,
					},
				},
			},
			STATE_COMPOSE_ERR: &lib.State{
				Transitions: lib.Transitions{
					TR_COMPOSE_WRITE: writeTr,
				},
			},
			STATE_COMPOSE_REVIEW_MAIL: &lib.State{
				Transitions: lib.Transitions{
					TR_COMPOSE_WRITE: writeTr,
					TR_COMPOSE_SET_ERR: &lib.Transition{
						Target: STATE_COMPOSE_ERR,
					},
					TR_COMPOSE_SEND: &lib.Transition{
						Target: STATE_COMPOSE_REVIEW_MAIL,
					},
				},
			},
		},
	)
}
