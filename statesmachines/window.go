package statesmachines

import (
	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/widgets"
)

const (
	STATE_SHOW_TAB lib.StateType      = "SHOW_TAB"
	TR_OPEN_TAB    lib.TransitionType = "OPEN_TAB"
	TR_CLOSE_TAB   lib.TransitionType = "CLOSE_TAB"
	TR_NEXT_TAB    lib.TransitionType = "NEXT_TAB"
	TR_PREV_TAB    lib.TransitionType = "PREV_TAB"
	TR_CLOSE_APP   lib.TransitionType = "CLOSE_APP"

	STATE_WRITE_CMD  lib.StateType      = "WRITE_CMD"
	TR_START_WRITING lib.TransitionType = "START_WRITING"
	TR_CANCEL        lib.TransitionType = "CANCEL"
	TR_VALIDATE      lib.TransitionType = "VALIDATE"
)

type Tab struct {
	Content widgets.Widget
	Title   string
}

type WindowMachineCtx struct {
	Tabs        []*Tab
	SelectedTab int
}

func NewWindowMachine() *lib.Machine {
	return lib.NewMachine(
		&WindowMachineCtx{
			Tabs:        make([]*Tab, 0),
			SelectedTab: 0,
		},
		STATE_SHOW_TAB,
		lib.States{
			STATE_SHOW_TAB: &lib.State{
				Transitions: lib.Transitions{
					TR_CLOSE_APP: &lib.Transition{
						Target: STATE_SHOW_TAB,
					},
					TR_OPEN_TAB: &lib.Transition{
						Target: STATE_SHOW_TAB,
						Action: func(c interface{}, ev *lib.Event) {
							wmc := c.(*WindowMachineCtx)
							newtab := ev.Payload.(*Tab)
							wmc.Tabs = append(wmc.Tabs, newtab)
							wmc.SelectedTab = len(wmc.Tabs) - 1
						},
					},
					TR_CLOSE_TAB: &lib.Transition{
						Target: STATE_SHOW_TAB,
						Action: func(c interface{}, ev *lib.Event) {
							wmc := c.(*WindowMachineCtx)
							if len(wmc.Tabs) == 1 {
								return
							}
							i := wmc.SelectedTab
							wmc.Tabs = append(wmc.Tabs[:i], wmc.Tabs[i+1:]...)
							wmc.SelectedTab = 0
						},
					},
					TR_NEXT_TAB: &lib.Transition{
						Target: STATE_SHOW_TAB,
						Action: func(c interface{}, ev *lib.Event) {
							wmc := c.(*WindowMachineCtx)
							next := wmc.SelectedTab + 1
							if next >= len(wmc.Tabs) {
								next = 0
							}
							wmc.SelectedTab = next
						},
					},
					TR_PREV_TAB:      &lib.Transition{Target: STATE_SHOW_TAB},
					TR_START_WRITING: &lib.Transition{Target: STATE_WRITE_CMD},
				},
			},
			STATE_WRITE_CMD: &lib.State{
				Transitions: lib.Transitions{
					TR_CANCEL:   &lib.Transition{Target: STATE_SHOW_TAB},
					TR_VALIDATE: &lib.Transition{Target: STATE_SHOW_TAB},
				},
			},
		},
	)
}
