package statesmachines

import (
	"github.com/stregouet/nuntius/lib"
)

const (
	STATE_STATUS_SHOW_MESSAGE lib.StateType      = "SHOW_MESSAGE"
	STATE_STATUS_WRITE_CMD    lib.StateType      = "WRITE_CMD"
	TR_STATUS_START_WRITING   lib.TransitionType = "START_WRITING"
	TR_STATUS_CANCEL          lib.TransitionType = "CANCEL"
	TR_STATUS_VALIDATE        lib.TransitionType = "VALIDATE"
	TR_STATUS_WRITE_CHAR      lib.TransitionType = "WRITE_CHAR"
	TR_STATUS_MOVE_CURSOR     lib.TransitionType = "MOVE_CURSOR"
	TR_STATUS_RM_CHAR         lib.TransitionType = "REMOVE_CHAR"
	TR_STATUS_RM_WORD         lib.TransitionType = "REMOVE_WORD"
	TR_STATUS_BROWSE_HISTORY  lib.TransitionType = "TR_STATUS_BROWSE_HISTORY"
)

type StatusMachineCtx struct {
	CursorPos    int
	WriteContent []rune
	History      []string
	HistoryIdx   int
}

func NewStatusMachine() *lib.Machine {
	reset := func(state *StatusMachineCtx) {
		state.CursorPos = 0
		state.WriteContent = []rune{}
		state.HistoryIdx = -1
	}

	return lib.NewMachine(
		&StatusMachineCtx{0, []rune{}, []string{}, -1},
		STATE_STATUS_SHOW_MESSAGE,
		lib.States{
			STATE_STATUS_SHOW_MESSAGE: &lib.State{
				Transitions: lib.Transitions{
					TR_STATUS_START_WRITING: &lib.Transition{
						Target: STATE_STATUS_WRITE_CMD,
					},
				},
			},
			STATE_STATUS_WRITE_CMD: &lib.State{
				Transitions: lib.Transitions{
					TR_STATUS_CANCEL: &lib.Transition{
						Target: STATE_STATUS_SHOW_MESSAGE,
						Action: func(c interface{}, ev *lib.Event) {
							state := c.(*StatusMachineCtx)
							reset(state)
						},
					},
					TR_STATUS_VALIDATE: &lib.Transition{
						Target: STATE_STATUS_SHOW_MESSAGE,
						Action: func(c interface{}, ev *lib.Event) {
							state := c.(*StatusMachineCtx)
							state.History = append(state.History, string(state.WriteContent))
							reset(state)
						},
					},

					TR_STATUS_WRITE_CHAR: &lib.Transition{
						Target: STATE_STATUS_WRITE_CMD,
						Action: func(c interface{}, ev *lib.Event) {
							state := c.(*StatusMachineCtx)
							char := ev.Payload.(rune)
							state.WriteContent = lib.InsertRune(state.WriteContent, state.CursorPos, char)
							state.CursorPos++
						},
					},
					TR_STATUS_RM_WORD: &lib.Transition{
						Target: STATE_STATUS_WRITE_CMD,
						Action: func(c interface{}, ev *lib.Event) {
							state := c.(*StatusMachineCtx)
							if state.CursorPos <= 1 {
								return
							}
							newcontent, dec := lib.RemoveRuneWordBackward(state.WriteContent, state.CursorPos)
							state.CursorPos -= dec
							state.WriteContent = newcontent
						},
					},
					TR_STATUS_RM_CHAR: &lib.Transition{
						Target: STATE_STATUS_WRITE_CMD,
						Action: func(c interface{}, ev *lib.Event) {
							state := c.(*StatusMachineCtx)
							if state.CursorPos <= 1 {
								return
							}
							state.WriteContent = lib.RemoveRuneBackward(state.WriteContent, state.CursorPos)
							state.CursorPos--
						},
					},
					TR_STATUS_MOVE_CURSOR: &lib.Transition{
						Target: STATE_STATUS_WRITE_CMD,
						Action: func(c interface{}, ev *lib.Event) {
							state := c.(*StatusMachineCtx)
							mv := ev.Payload.(int)
							state.CursorPos += mv
							if state.CursorPos < 1 {
								state.CursorPos = 1
							} else if state.CursorPos > len(state.WriteContent) {
								state.CursorPos = len(state.WriteContent)
							}
						},
					},
					TR_STATUS_BROWSE_HISTORY: &lib.Transition{
						Target: STATE_STATUS_WRITE_CMD,
						Action: func(c interface{}, ev *lib.Event) {
							state := c.(*StatusMachineCtx)
							if len(state.History) == 0 {
								return
							}
							mv := ev.Payload.(int)
							state.HistoryIdx += mv
							if state.HistoryIdx < 0 {
								state.HistoryIdx = len(state.History) - 1
							} else if state.HistoryIdx >= len(state.History) {
								state.HistoryIdx = 0
							}
							state.WriteContent = []rune(state.History[state.HistoryIdx])
							state.CursorPos = len(state.WriteContent)
						},
					},
				},
			},
		},
	)
}
