package widgets

import (
  "log"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
)


type Window struct {
  app *views.Application
  logger *log.Logger

	views.Panel
}


func NewWindow(a *views.Application, l *log.Logger) Window {
  return Window{app: a, logger: l}
}

func (w *Window) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyCtrlL:
			w.app.Refresh()
			return true
		case tcell.KeyRune:
			switch ev.Rune() {
			case 'Q', 'q':
				w.app.Quit()
				return true
      }
    }
  }
  w.logger.Printf("window handle event fallback %v", ev)
  return w.Panel.HandleEvent(ev)
}
