package statesmachines

import (
	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/models"
)

const (
	STATE_MAIL_PARTS     lib.StateType      = "SHOW_MAIL"
	TR_MAIL_PARTS_UP     lib.TransitionType = "MAIL_PARTS_UP"
	TR_MAIL_PARTS_DOWN   lib.TransitionType = "MAIL_PARTS_DOWN"
	TR_SET_SELECTED_PART lib.TransitionType = "SET_SELECTED_PART"
	TR_SELECT_PART       lib.TransitionType = "SELECT_PART"
)

type MailPartsMachineCtx struct {
	Parts    []*models.BodyPart
	Selected int
}

func NewMailPartsMachine(parts []*models.BodyPart) *lib.Machine {
	min := 1
	for i, p := range parts {
		if p.MIMEType != "multipart" {
			min = i + 1
			break
		}
	}
	return lib.NewMachine(
		&MailPartsMachineCtx{
			Parts:    parts,
			Selected: min,
		},
		STATE_MAIL_PARTS,
		lib.States{
			STATE_MAIL_PARTS: &lib.State{
				Transitions: lib.Transitions{
					TR_SELECT_PART: &lib.Transition{
						Target: STATE_MAIL_PARTS,
					},
					TR_MAIL_PARTS_DOWN: &lib.Transition{
						Target: STATE_MAIL_PARTS,
						Action: func(c interface{}, ev *lib.Event) {
							state := c.(*MailPartsMachineCtx)
							next := state.Selected
							for {
								next++
								if next > len(state.Parts) {
									next = len(state.Parts)
									break
								}
								if state.Parts[next-1].MIMEType != "multipart" {
									break
								}

							}
							state.Selected = next
						},
					},
					TR_MAIL_PARTS_UP: &lib.Transition{
						Target: STATE_MAIL_PARTS,
						Action: func(c interface{}, ev *lib.Event) {
							state := c.(*MailPartsMachineCtx)
							next := state.Selected
							for {
								next--
								if next < min {
									next = min
									break
								}
								if state.Parts[next-1].MIMEType != "multipart" {
									break
								}
							}
							state.Selected = next
						},
					},
					TR_SET_SELECTED_PART: &lib.Transition{
						Target: STATE_MAIL_PARTS,
						Action: func(c interface{}, ev *lib.Event) {
							state := c.(*MailPartsMachineCtx)
							next := ev.Payload.(int)
							state.Selected = next
						},
					},
				},
			},
		},
	)
}
