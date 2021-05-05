package widgets

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"

	"github.com/stregouet/nuntius/ui"
)

type Window struct {
	appCh  chan (ui.AppEvent)
	logger *log.Logger

	views.Panel
}

func NewWindow(ch chan (ui.AppEvent), l *log.Logger) *Window {
	return &Window{appCh: ch, logger: l}
}

func (w *Window) sendAppEvent(ev ui.AppEvent) {
	w.appCh <- ev
}

func (w *Window) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyRune:
			switch ev.Rune() {
			case 'Q', 'q':
				w.sendAppEvent(ui.QUIT_EVENT)
				return true
			}
		}
	}
	return w.Panel.HandleEvent(ev)
}
