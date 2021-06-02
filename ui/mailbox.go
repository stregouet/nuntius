package ui



import (
	"github.com/gdamore/tcell/v2"

	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/models"
	"github.com/stregouet/nuntius/widgets"
)

const (
	STATE_LOAD_MBOX lib.StateType      = "LOAD_MBOX"
	STATE_SHOW_MBOX lib.StateType      = "SHOW_MBOX"
	TR_SET_THREADS     lib.TransitionType = "SET_THREADS"
	TR_REFRESH_MBOX     lib.TransitionType = "REFRESH_MBOX"
)

type MailboxMachineCtx struct {
	threads   []*models.Thread
	selected int
}

func buildMailboxMachine() *lib.Machine {
    c := &MailboxMachineCtx{
        threads: make([]*models.Thread, 0),
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

type MailboxView struct {
    machine *lib.Machine
	*widgets.ListWidget
}

func NewMailboxView(accountName string, onSelect func(accname string, t *models.Thread)) *MailboxView {
	l := widgets.NewList()
	return &MailboxView{
		machine:     buildMailboxMachine(),
		ListWidget:  l,
	}
}


func (mv *MailboxView) SetThreads(threads []*models.Thread) {
	mv.machine.Send(&lib.Event{TR_SET_THREADS, threads})
	mv.ClearLines()
	for _, t := range threads {
		mv.AddLine(t)
	}
	mv.AskRedraw()
}

func (mv *MailboxView) Draw() {
	if mv.machine.Current == STATE_LOAD_MBOX {
		style := tcell.StyleDefault
		for i, c := range "loading..." {
			mv.SetContent(i, 0, c, nil, style)
		}
	} else {
		mv.ListWidget.Draw()
	}
}

func (mv *MailboxView) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyRune:
			switch ev.Rune() {
			case 'Q', 'q':
				App.Stop()
				return true
			}
		}
	}
	return mv.ListWidget.HandleEvent(ev)
}
